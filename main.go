package main

import (
	"google.golang.org/protobuf/proto"
	"net/http"
	"time"
	"ws/consts"
	"ws/pb"
	"ws/ws"
)

// We'll need to define an Upgrader
// this will require a Read and Writer buffer size

func main() {
	go ws.Manager.Run()

	http.HandleFunc("/ws", ws.Endpoint)
	//http.HandleFunc("/designate", Designate)

	//go broadcastTimer()
	if err := http.ListenAndServe(":7777", nil); err != nil {
		return
	}

}

func Designate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	//msg := r.URL.Query().Get("msg")

	for client, _ := range ws.Manager.Clients {
		if client.Token == token {
			// TODO
			break
		}
	}
}

func broadcastTimer() {

	for {
		timer := time.NewTimer(3 * time.Second)
		select {
		case <-timer.C:
			heartbeatReq := &pb.HeartbeatReq{
				Token: "test",
			}

			heartbeatReqProto, _ := proto.Marshal(heartbeatReq)

			payload := &pb.Payload{
				Type: consts.WsHeartbeatReq,
				Data: heartbeatReqProto,
			}

			payloadProto, _ := proto.Marshal(payload)
			ws.Manager.Broadcast <- payloadProto
		}
	}
}
