package logic

import (
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"ws/consts"
	"ws/pb"
)

func HeartbeatReq(conn *websocket.Conn, data []byte) {
	heartbeatReq := &pb.HeartbeatReq{}
	if err := proto.Unmarshal(data, heartbeatReq); err != nil {
		return
	}

	heartbeatResp := &pb.HeartBeatResp{}
	heartbeatResp.Code = 200
	heartbeatRespProto, _ := proto.Marshal(heartbeatResp)

	payload := &pb.Payload{
		Type: consts.WsHeartbeatResp,
		Data: heartbeatRespProto,
	}

	payloadProto, _ := proto.Marshal(payload)

	conn.WriteMessage(websocket.PingMessage, payloadProto)
}
