package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"net/http"
	"time"
	"ws/consts"
	"ws/logic"
	"ws/pb"
	"strings"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

type ContainerPackage struct {
	Msg []byte
	ContainerId string
}

type Package struct {
	Msg    []byte
	Client *Client
}

type ClientManager struct {
	Register   chan *Client
	Unregister chan *Client
	ContainerBroadcast  chan *ContainerPackage
	Designator chan *Package
	Containers map[string]map[*Client]struct{}
}

type Client struct {
	Conn  *websocket.Conn
	ContainerId string
	Send  chan []byte
	Token string
	TTL   int8
}

var Manager = &ClientManager{
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	ContainerBroadcast:  make(chan *ContainerPackage),
	Designator: make(chan *Package),
	Containers: make(map[string]map[*Client]struct{}),
}

func (manager *ClientManager) Run() {
	for {
		select {
		case client := <-manager.Register:
			if manager.Containers[client.ContainerId] == nil {
				manager.Containers[client.ContainerId] = make(map[*Client]struct{})
			}

			manager.Containers[client.ContainerId][client] = struct{}{}
			for containerId, container := range manager.Containers {
				title := fmt.Sprintf("@@@@@@@@@@@@@@@@@@@@ROOM#%s@@@@@@@@@@@@@@@@@@@@", containerId)
				fmt.Printf("%s\n", title)
				for client, _ := range container {
					fmt.Printf("#%s#: %v\n", client.Token, client)
				}
				fmt.Printf("%s\n\n", strings.Repeat("@", len(title)))
			}
		case client := <-manager.Unregister:
			client.Conn.Close()
			close(client.Send)
			delete(manager.Containers[client.ContainerId], client)
			if len(manager.Containers[client.ContainerId]) == 0 {
				delete(manager.Containers, client.ContainerId)
			}

			for containerId, container := range manager.Containers {
				title := fmt.Sprintf("@@@@@@@@@@@@@@@@@@@@ROOM#%s@@@@@@@@@@@@@@@@@@@@", containerId)
				fmt.Printf("%s\n", title)
				for client, _ := range container {
					fmt.Printf("#%s#: %v\n", client.Token, client)
				}
				fmt.Printf("%s\n\n", strings.Repeat("@", len(title)))
			}
		case containerPackage := <-manager.ContainerBroadcast:
			for client, _ := range manager.Containers[containerPackage.ContainerId] {
				select {
				case client.Send <- containerPackage.Msg:
					//default:
					//	close(client.Send)
					//	delete(Manager.Clients, client)
				}

			}
		case p := <-manager.Designator:
			p.Client.Send <- p.Msg
		}

	}
}
func Endpoint(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token == "" {
		fmt.Println("error argument: need token in headers.")
		return
	}

	containerId := r.Header.Get("container_id")
	if containerId == "" {
		fmt.Println("error argument: need room in headers.")
		return
	}
	
	ws, _ := upgrade.Upgrade(w, r, nil)
	client := &Client{
		Conn:  ws,
		ContainerId: containerId,
		Token: r.Header.Get("token"),
		Send:  make(chan []byte),
		TTL:   consts.ClientDefaultTTL,
	}
	Manager.Register <- client

	go client.Read()
	go client.Write()
}

func (c *Client) Write() {
	defer c.Conn.Close()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.BinaryMessage, msg)
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
				c.Conn.Close()
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
		proto.Unmarshal(msg, payload)

		switch payload.Type {
		case consts.WsHeartbeatReq:
			c.TTL = consts.ClientDefaultTTL
			logic.HeartbeatReq(c.Conn, payload.Data)
		}

	}

}
