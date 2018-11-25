package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/xedinaska/tcpcommunicationv2/clientapp/command"
	"github.com/xedinaska/tcpcommunicationv2/model"
)

//Executor describes interface of server commands
type Executor interface {
	Exec()
}

func main() {
	serverHostPtr := flag.String("server-host", "localhost", "use -server-host to provide server host for client app (localhost by default)")
	serverPortPtr := flag.Int("server-port", 3333, "use -server-port to provide server port for client app (3333 by default)")
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	server := fmt.Sprintf("%s:%d", *serverHostPtr, *serverPortPtr)

	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatalf("host=%s,port=%d; failed to connect to provided server: `%s`", *serverHostPtr, *serverPortPtr, err.Error())
	}

	defer conn.Close()

	log.Infof("client app successfully connected to server %s", server)

	//listen for server messages in separate routine
	go serve(conn)

	//reading commands from stdin
	for in := range stdIN(os.Stdin) {
		msg := &model.Message{}
		if err := json.Unmarshal([]byte(in), msg); err != nil {
			log.Errorf("failed to read input: `%s`", err.Error())
			continue
		}

		if err := msg.Validate(); err != nil {
			log.Error(err.Error())
			continue
		}

		if msg.Type == "command" {
			cmd := msg.Payload.(string)
			executor(cmd, conn).Exec()
			continue
		}

		if msg.Type == "message" {
			log.Debugf("should send message to another client: `%v`", msg.Payload)
			conn.Write([]byte(in))
		}
	}
}

//serve is used to handle server commands
func serve(conn net.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			log.Errorf("failed to read server message: `%s`", err.Error())
			break
		}

		msg := &model.Message{}
		if err := json.Unmarshal(buff[:n], &msg); err != nil {
			log.Errorf("failed to unmarshal server response: `%s`", err.Error())
			continue
		}

		if err := msg.Validate(); err != nil {
			log.Error(err.Error())
			continue
		}

		if msg.Type == "message" {
			payload := msg.Payload.(map[string]interface{})
			log.Debugf("received message: `%s`", payload["text"])
			continue
		}

		if msg.Type == "clients" {
			data := msg.Payload.(map[string]interface{})

			//convert to clients payload
			payload, err := model.ReadClientsPayload(data)
			if err != nil {
				log.Errorf("failed to read clients payload: `%s`", err.Error())
				continue
			}

			log.Debugf("connected clients: ")
			for _, cc := range payload.Clients {
				log.Debugf("{ID: %s, IP: %s}", cc.ID, cc.Address)
			}
		}

		if msg.Type == "command" && (msg.Payload.(string) == command.STOP) {
			log.Info("STOP signal received; stopping client..")
			os.Exit(0)
		}
	}
}

//stdIN used to run the goroutine that reading stdin messages
func stdIN(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			lines <- scan.Text()
		}
	}()
	return lines
}

//executor is used to find required executor by provided command name
func executor(cmd string, conn net.Conn) Executor {
	if cmd == command.STOP {
		return &command.Stop{}
	}

	if cmd == command.LIST {
		return &command.List{Conn: conn}
	}

	return &command.Default{}
}
