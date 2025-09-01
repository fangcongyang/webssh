package webssh

import (
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

const (
	MsgData   = "1"
	MsgResize = "2"
)

type Turn struct {
	StdinPipe io.WriteCloser
	Session   *ssh.Session
	WsConn    *websocket.Conn
	Recorder  *Recorder
}

func NewTurn(wsConn *websocket.Conn) *Turn {
	return &Turn{WsConn: wsConn}
}

func (t *Turn) initSshConn(sshClient *ssh.Client, rec *Recorder) {
	sess, err := sshClient.NewSession()
	if err != nil {
		t.WsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}
	t.Session = sess

	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		t.WsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}
	t.StdinPipe = stdinPipe
	sess.Stdout = t
	sess.Stderr = t

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sess.RequestPty("xterm", 150, 30, modes); err != nil {
		t.WsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}
	if err := sess.Shell(); err != nil {
		t.WsConn.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
	}

	if rec != nil {
		t.Recorder = rec
		t.Recorder.Lock()
		t.Recorder.WriteHeader(30, 150)
		t.Recorder.Unlock()
	}

	//
	go func() {
		err := t.SessionWait()
		if err != nil {
			println(fmt.Errorf("session错误: %v", err))
		}
	}()
}

func (t *Turn) Write(p []byte) (n int, err error) {
	writer, err := t.WsConn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	if t.Recorder != nil {
		t.Recorder.Lock()
		t.Recorder.WriteData(OutPutType, string(p))
		t.Recorder.Unlock()
	}
	return writer.Write(p)
}

func (t *Turn) Handler(args *Command) error {
	switch args.OperType {
	case MsgResize:
		if args.Cols >= 0 && args.Rows > 0 {
			if err := t.Session.WindowChange(args.Rows, args.Cols); err != nil {
				return fmt.Errorf("ssh pty resize windows err:%s", err)
			}
		}
	case MsgData:
		if _, err := t.StdinPipe.Write([]byte(args.Command)); err != nil {
			return fmt.Errorf("StdinPipe write err:%s", err)
		}
	}
	return nil
}

func (t *Turn) Close() {
	if t.Session != nil {
		t.Session.Close()
	}
}

func (t *Turn) SessionWait() error {
	if err := t.Session.Wait(); err != nil {
		return err
	}
	return nil
}

func decode(p []byte) []byte {
	decodeString, _ := base64.StdEncoding.DecodeString(string(p))
	return decodeString
}

func encode(p []byte) []byte {
	encodeToString := base64.StdEncoding.EncodeToString(p)
	return []byte(encodeToString)
}

type Command struct {
	OperType string `json:"operType"`
	Cols     int    `json:"cols"`
	Rows     int    `json:"rows"`
	Command  string `json:"command"`
}
