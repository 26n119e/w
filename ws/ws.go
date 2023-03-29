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
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

type Package struct {
	Msg    []byte
	Client *Client
}

type ClientManager struct {
	Clients    map[*Client]struct{}
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	Designator chan *Package
}

type Client struct {
	Conn  *websocket.Conn
	Send  chan []byte
	Token string
	TTL   int8
}

var Manager = &ClientManager{
	Clients:    make(map[*Client]struct{}),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Broadcast:  make(chan []byte),
	Designator: make(chan *Package),
}

func (manager *ClientManager) Run() {
	for {
		select {
		case client := <-manager.Register:
			manager.Clients[client] = struct{}{}
			fmt.Printf("clients: %v\n", Manager.Clients)
		case client := <-manager.Unregister:
			client.Conn.Close()
			close(client.Send)
			delete(manager.Clients, client)
			fmt.Printf("clients: %v\n", Manager.Clients)
		case msg := <-manager.Broadcast:
			for client, _ := range manager.Clients {
				select {
				case client.Send <- msg:
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
	ws, _ := upgrade.Upgrade(w, r, nil)
	client := &Client{
		Conn:  ws,
		Token: r.Header.Get("token"),
		Send:  make(chan []byte),
		TTL:   consts.ClientDefaultTLL,
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
			fmt.Printf("TTL: %v\n", c.TTL)
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
			c.TTL = consts.ClientDefaultTLL
			logic.HeartbeatReq(c.Conn, payload.Data)
		}

	}

}
