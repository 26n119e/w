package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"ws/consts"
	"ws/ws"
)

// We'll need to define an Upgrader
// this will require a Read and Writer buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

func main() {
	go ws.Manager.Run()

	http.HandleFunc("/ws", endpoint)
	//http.HandleFunc("/designate", designate)
	http.HandleFunc("/container_broadcast", broadcast2container)

	//go broadcastTimer()
	if err := http.ListenAndServe(":7777", nil); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

}

func designate(w http.ResponseWriter, r *http.Request) {
	// token := r.URL.Query().Get("token")
	// msg := r.URL.Query().Get("msg")

	// for client, _ := range ws.Manager.Clients {
	// 	if client.Token == token {
	// 		break
	// 	}
	// }
}

func endpoint(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		fmt.Println("error argument: need token in headers.")
		return
	}

	containerId := r.URL.Query().Get("container_id")
	if containerId == "" {
		fmt.Println("error argument: need room in headers.")
		return
	}

	conn, _ := upgrader.Upgrade(w, r, nil)
	client := &ws.Client{
		Conn:        conn,
		ContainerId: containerId,
		Token:       token,
		Send:        make(chan []byte),
		TTL:         consts.ClientDefaultTTL,
	}
	ws.Manager.Register <- client

	go client.Read()
	go client.Write()
}

func broadcast2container(_ http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	if msg == "" {
		fmt.Printf("%s\n", "error argument: need msg in headers.")
		return
	}
	containerId := r.URL.Query().Get("container_id")
	if containerId == "" {
		fmt.Printf("%s\n", "error argument: need container_id in headers.")
		return
	}

	// example, unused protobuf
	containerPackage := &ws.ContainerPackage{
		ContainerId: containerId,
		Msg:         []byte(msg),
	}

	ws.Manager.ContainerBroadcast <- containerPackage
}
