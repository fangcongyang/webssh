package webssh

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"net/http"
)


func WsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *Connection
		data   []byte
	)
	// 完成ws协议的握手操作
	// Upgrade:websocket
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	if conn, err = InitConnection(wsConn); err != nil {
		goto ERR
	}
	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		if data != nil {
			var sshMes SshMessage
			err := json.Unmarshal(data, &sshMes)
			if err != nil {
				conn.WriteMessage([]byte("JsonToMapDemo err: " + string(data)))
			}
			switch sshMes.MesType {
			case "1":
				handle := NewWebSSH(sshMes.Config)
				c, r := handle.ServeConn(wsConn)
				conn.turn.initSshConn(c, r)
			case "2":
				conn.turn.Handler(sshMes.Command)
			}
		}
	}

ERR:
	conn.Close()

}

type SshMessage struct {
	Command *Command `json:"data"`
	SshId string `json:"sshId"`
	MesType string `json:"mesType"`
	Config *WebSSHConfig `json:"config"`
}