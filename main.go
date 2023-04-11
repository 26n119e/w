package main

import (
	"net/http"
	"ws/ws"
	"fmt"
)

// We'll need to define an Upgrader
// this will require a Read and Writer buffer size

func main() {
	go ws.Manager.Run()

	http.HandleFunc("/ws", ws.Endpoint)
	//http.HandleFunc("/designate", Designate)
	http.HandleFunc("/container_broadcast", broadcast2container)

	//go broadcastTimer()
	if err := http.ListenAndServe(":7777", nil); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

}

func Designate(w http.ResponseWriter, r *http.Request) {
	// token := r.URL.Query().Get("token")
	// msg := r.URL.Query().Get("msg")

	// for client, _ := range ws.Manager.Clients {
	// 	if client.Token == token {
	// 		// TODO
	// 		break
	// 	}
	// }
}

func broadcast2container(w http.ResponseWriter, r *http.Request) {
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

	// example, unuse protobuf
	containerPackage := &ws.ContainerPackage{
		ContainerId: containerId,
		Msg: []byte(msg),
	}

	ws.Manager.ContainerBroadcast <- containerPackage
}
