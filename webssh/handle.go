package webssh

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/gorilla/websocket"
)

type WebSSHConfig struct {
	Record    bool
	RecPath   string
	HostAddr  string
	Username  string
	Password  string
	AuthModel AuthModel
	PkPath    string
}

type WebSSH struct {
	*WebSSHConfig
}

func NewWebSSH(conf *WebSSHConfig) *WebSSH {
	return &WebSSH{
		WebSSHConfig: conf,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (w WebSSH) ServeConn(wsConn *websocket.Conn) (c *ssh.Client, r *Recorder) {
	var config *SSHClientConfig
	switch w.AuthModel {

	case PASSWORD:
		config = SSHClientConfigPassword(
			w.HostAddr,
			w.Username,
			w.Password,
		)
	case PUBLICKEY:
		config = SSHClientConfigPulicKey(
			w.HostAddr,
			w.Username,
			w.PkPath,
		)
	}

	client, err := NewSSHClient(config)
	if err != nil {
		wsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}

	var recorder *Recorder
	if w.Record {
		os.MkdirAll(w.RecPath, 0766)
		fileName := path.Join(w.RecPath, fmt.Sprintf("%s_%s_%s.cast", w.HostAddr, w.Username, time.Now().Format("20060102_150405")))
		var f *os.File
		f, err = os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0766)
		if err != nil {
			println(fmt.Errorf("初始化日志文件失败: %v", err))
		}
		defer f.Close()
		recorder = NewRecorder(f)
	}

	if err != nil {
		wsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}
	return client, recorder
}
