package command

import (
	"encoding/json"
	"github.com/xedinaska/tcpcommunicationv2/model"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
)

const (
	STOP = "STOP"
	LIST = "CLIENTS_LIST"
)

type Stop struct{}

func (h *Stop) Exec() {
	log.Info("STOP signal received; stopping client..")
	os.Exit(0)
}

type Default struct{}

func (h *Default) Exec() {
	log.Errorf("failed to find executor, use default")
}

type List struct {
	Conn net.Conn
}

func (h *List) Exec() {
	msg := &model.Message{
		Type:    "command",
		Payload: "CLIENTS_LIST",
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("failed to marshal server command: `%s`", err.Error())
	}

	h.Conn.Write(b)
}
