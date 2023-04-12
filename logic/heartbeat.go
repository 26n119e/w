package logic

import (
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"ws/consts"
	"ws/pb"
)

func SendHeartbeatResp(conn *websocket.Conn) {
	heartbeatRespProto, _ := proto.Marshal(&pb.HeartBeatResp{Code: 200})

	payload := &pb.Payload{
		Type: consts.WsHeartbeatResp,
		Data: heartbeatRespProto,
	}

	payloadProto, _ := proto.Marshal(payload)

	conn.WriteMessage(websocket.BinaryMessage, payloadProto)
}
