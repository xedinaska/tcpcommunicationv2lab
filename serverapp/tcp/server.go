package tcp

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/xedinaska/tcpcommunicationv2/model"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

//NewServer returns new tcp.Server instance with initialized fields. It accepts host/port as input params
func NewServer(host string, port int) *Server {
	return &Server{
		host:     host,
		port:     port,
		stopChan: make(chan os.Signal),
		clients:  make(map[string]*model.Client, 0),
	}
}

//Server represents TCP server & contains handle methods
type Server struct {
	host     string
	port     int
	clients  map[string]*model.Client
	stopChan chan os.Signal
	listener net.Listener
}

//Start will start TCP listener
func (s *Server) Start() (err error) {
	//accept TCP connections on provided port
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		log.Printf("[ERROR]host=%s,port=%d; unable to start listener: `%s`", s.host, s.port, err.Error())
		return err
	}

	//listen for os signals to (SIGTERM / SIGINT) for graceful shutdown
	signal.Notify(s.stopChan, syscall.SIGTERM)
	signal.Notify(s.stopChan, syscall.SIGINT)

	go s.shutdown() //gracefully handle server shutdown (close connections, free resources / etc)

	//infinite loop in separate goroutine to handle incoming connections
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("[ERROR]failed to accept incoming connection: `%s`", err.Error())
			}
			c := s.addClient(conn)
			log.Printf("[DEBUG]incoming connection: %s, clients connected: %d", conn.RemoteAddr(), len(s.clients))

			go s.handle(c)
		}
	}()

	return nil
}

//handle used to handle client connection (listen messages, disconnect, etc..)
func (s *Server) handle(c *model.Client) {
	buf := make([]byte, 1024)
	r := bufio.NewReader(c.Conn)

LOOP:
	for {
		n, err := r.Read(buf)
		switch err {
		case nil:
			msg := &model.Message{}
			if readerr := json.Unmarshal(buf[:n], msg); readerr != nil {
				log.Printf("[ERROR] failed to unmarshal client message: `%s`", readerr.Error())
				continue
			}

			log.Printf("[DEBUG][%s] received client message: `%s`:", c.Address, msg)

			//s.messages = append(s.messages, msg.Payload.Message)
			if msg.Type == "command" && msg.Payload.(string) == "CLIENTS_LIST" {
				clients := make([]*model.Client, 0)
				for _, cc := range s.clients {
					clients = append(clients, cc)
				}

				answ := &model.Message{
					Type: "clients",
					Payload: &model.ClientsPayload{
						Clients: clients,
					},
				}

				b, encerr := json.Marshal(answ)
				if encerr != nil {
					log.Errorf("failed to marshal client message: `%s`", encerr.Error())
					continue
				}

				if serr := c.Send(b); serr != nil {
					log.Errorf("failed to sent message to client: `%s`", serr.Error())
				}
				continue
			}

			if msg.Type == "message" {
				payload, err := model.ReadMessagePayload(msg.Payload.(map[string]interface{}))
				if err != nil {
					log.Errorf("failed to read message: `%s`", err.Error())
					continue
				}

				client, err := s.getClient(payload.ID, payload.IP)
				if err != nil {
					log.Errorf(err.Error())
					continue
				}

				client.Send(buf[:n])
			}

		case io.EOF:
			log.Printf("[DEBUG]EOF, disconnecting client %s..", c.Address)

			//close TCP conn / etc
			c.Disconnect()

			//remove client from clients slice
			delete(s.clients, c.ID)

			log.Printf("[DEBUG]..done, clients connected: %d", len(s.clients))
			break LOOP
		default:
			c.Disconnect()
			log.Printf("[ERROR]failed to read input message: `%s`", err.Error())
			break
		}
	}
}

//addClient is used to generate init client by provided conn, generate unique id and add client to connected clients slice
func (s *Server) addClient(conn net.Conn) *model.Client {
	c := &model.Client{
		ID:      hex.EncodeToString(md5.New().Sum([]byte(conn.RemoteAddr().String()))),
		Conn:    conn,
		Address: conn.RemoteAddr().String(),
	}

	s.clients[c.ID] = c
	return c
}

//getClient will return client by id or ip (id - from map, ip - go through map of connected clients)
//in case then clients isn't found it'll return an error
func (s *Server) getClient(id, ip string) (*model.Client, error) {
	if id != "" {
		c, ok := s.clients[id]
		if !ok {
			return nil, fmt.Errorf("failed to find client #%s", id)
		}
		return c, nil
	}

	for _, cc := range s.clients {
		if cc.Address == ip {
			return cc, nil
		}
	}

	return nil, fmt.Errorf("failed to find client #%s", ip)
}

//shutdown used to disconnect all connected clients after shutdown server signal && close TCP connection
func (s *Server) shutdown() {
	<-s.stopChan
	log.Printf("[INFO]received shutdown signal. Stopping %d clients & exit..", len(s.clients))

	for _, c := range s.clients {
		if err := c.Disconnect(); err != nil {
			log.Printf("[ERROR] failed to disconnect client %s: `%s`", c.Address, err.Error())
		}
	}

	log.Printf("[INFO]..done. Exit")

	s.listener.Close()
	os.Exit(0)
}
