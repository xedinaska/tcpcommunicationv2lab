package model

import (
	"net"
	"fmt"
	"encoding/json"
)

//Message used for communication: "client-server", "server-client", "STDIN-client".
//examples are available in README.md
type Message struct {
	Type string `json:"type"`
	Payload interface{} `json:"payload"`
}

//Validate is used to filter invalid commands
func (m *Message) Validate() error {
	if m.Type != "message" && m.Type != "command" && m.Type != "clients" {
		return fmt.Errorf("available types of messages: `message` or `command` - invalid type `%s` received", m.Type)
	}
	return nil
}

//MessagePayload will sent by client to server & from server to client then client wants to send smth to another one
type MessagePayload struct {
	ID   string `json:"id"`
	IP   string `json:"ip"`
	Text string `json:"text"`
}

//ClientsPayload will be sent to client from server then it asks "connected clients list"
type ClientsPayload struct {
	Clients []*Client
}

//Client describes connected client
type Client struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Conn    net.Conn `json:"-"`
}

//Disconnect is used to close connection / free resources / etc
func (c *Client) Disconnect() error {
	msg := &Message{
		Type: "command",
		Payload: "STOP",
	}
	b, _ := json.Marshal(msg)
	c.Conn.Write(b)
	return c.Conn.Close()
}

//Send is used to write message to client
func (c *Client) Send(msg []byte) error {
	_, err := c.Conn.Write(msg)
	return err
}

//ReadClientsPayload is used to convert map to payload
func ReadClientsPayload(data map[string]interface{}) (*ClientsPayload, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	clients := &ClientsPayload{}
	if err := json.Unmarshal(b, clients); err != nil {
		return nil, err
	}

	return clients, nil
}

//ReadMessagePayload is used to convert map to payload
func ReadMessagePayload(data map[string]interface{}) (*MessagePayload, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	message := &MessagePayload{}
	if err := json.Unmarshal(b, message); err != nil {
		return nil, err
	}

	return message, nil
}