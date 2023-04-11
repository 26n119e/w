package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
	"ws/consts"
	"ws/logic"
	"ws/pb"
)

type ContainerPackage struct {
	Msg         []byte
	ContainerId string
}

type Package struct {
	Msg    []byte
	Client *Client
}

type ClientManager struct {
	Register           chan *Client
	Unregister         chan *Client
	ContainerBroadcast chan *ContainerPackage
	Designator         chan *Package
	Containers         map[string]map[*Client]struct{}
}

type Client struct {
	Conn        *websocket.Conn
	ContainerId string
	Send        chan []byte
	Token       string
	TTL         int8
}

var Manager = &ClientManager{
	Register:           make(chan *Client),
	Unregister:         make(chan *Client),
	ContainerBroadcast: make(chan *ContainerPackage),
	Designator:         make(chan *Package),
	Containers:         make(map[string]map[*Client]struct{}),
}

func (cm *ClientManager) Run() {
	for {
		select {
		case client := <-cm.Register:
			if cm.Containers[client.ContainerId] == nil {
				cm.Containers[client.ContainerId] = make(map[*Client]struct{})
			}

			cm.Containers[client.ContainerId][client] = struct{}{}
			cm.inspector()
		case client := <-cm.Unregister:
			client.Conn.Close()
			close(client.Send)
			delete(cm.Containers[client.ContainerId], client)
			if len(cm.Containers[client.ContainerId]) == 0 {
				delete(cm.Containers, client.ContainerId)
			}
			cm.inspector()
		case containerPackage := <-cm.ContainerBroadcast:
			for client, _ := range cm.Containers[containerPackage.ContainerId] {
				client.Send <- containerPackage.Msg
			}
		case p := <-cm.Designator:
			p.Client.Send <- p.Msg
		}

	}
}

func (cm *ClientManager) inspector() {
	for containerId, container := range cm.Containers {
		title := fmt.Sprintf(strings.Repeat("@", 33)+"ROOM#%s"+strings.Repeat("@", 33), containerId)
		fmt.Printf("%s\n", title)
		for client, _ := range container {
			fmt.Printf("%v\n", client)
		}
		fmt.Printf("%s\n\n", strings.Repeat("@", len(title)))
	}
}

func (c *Client) Write() {
	defer c.Conn.Close()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			_ = c.Conn.WriteMessage(websocket.BinaryMessage, msg)
		}
	}
}

func (c *Client) timeout(ctx context.Context) {
	for {
		TTLTimer := time.NewTimer(1 * time.Second)
		select {
		case <-TTLTimer.C:
			c.TTL--
			// fmt.Printf("%s's TTL: %v\n", c.Token, c.TTL)
			if c.TTL <= 0 {
				_ = c.Conn.Close()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
	}()

	ctx, cancel := context.WithCancel(context.Background())

	go c.timeout(ctx)

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			cancel()
			break
		}

		payload := &pb.Payload{}
		if err := proto.Unmarshal(msg, payload); err != nil {
			return
		}

		switch payload.Type {
		case consts.WsHeartbeatReq:
			c.TTL = consts.ClientDefaultTTL
			logic.HeartbeatReq(c.Conn, payload.Data)
		}

	}

}
