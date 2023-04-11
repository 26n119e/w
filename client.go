package main

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

var (
	containerIdList = []string{"aaaaa", "bbbbb", "ccccc", "aaaaa"}
)

func main() {
	for i := 0; i < 100; i++ {
		url := fmt.Sprintf("ws://%s:%s/ws?container_id=%s&token=%s", "localhost", "7777", containerIdList[i%4], fetchUUID())
		c, _, _ := websocket.DefaultDialer.Dial(url, nil)
		defer c.Close()

		go func() {
			for {
				_, message, _ := c.ReadMessage()
				fmt.Printf("message: %s\n", message)
			}
		}()
	}
	select {}
}

func fetchUUID() string {
	return uuid.NewV4().String()
}
