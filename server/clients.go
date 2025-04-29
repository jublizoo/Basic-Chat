package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Error codes
const (
	SERVER_ERR = iota
	USER_DISCONNECTED
)

type ConnReq struct {
	from string
	to   string
}

type ConnRes struct {
	from     string
	to       string
	accepted bool
	ok       bool
	err      int
}

type DisconnReq struct {
	from string
	to   string
}

type ChatMsg struct {
	from string
	to   string
	body string
}

type Envelope struct {
	msg_type string
	payload  []byte
}

type Client struct {
	conn    *websocket.Conn
	user    string
	connRes chan ConnRes
	close   chan struct{}
}

type Client_list struct {
	users       map[string]*Client
	connections map[string]string // Maps connected users to their connections
	requests    chan interface{}
}

func initializeClients() *Client_list {
	clients := Client_list{}
	clients.users = make(map[string]*Client)
	clients.connections = make(map[string]string)
	clients.requests = make(chan interface{})
	return &clients
}

func unwrapEnvelope(envelope Envelope) interface{} {
	switch envelope.msg_type {
	case "Message":
		var msg ChatMsg
		json.Unmarshal(envelope.payload, &msg)
		return msg
	case "ConnReq":
		var msg ConnReq
		json.Unmarshal(envelope.payload, &msg)
		return msg
	case "ConnRes":
		var msg ConnRes
		json.Unmarshal(envelope.payload, &msg)
		return msg
	case "DisconnReq":
		var msg DisconnReq
		json.Unmarshal(envelope.payload, &msg)
		return msg
	default:
		panic("Bad envelope type")
	}
}

func (client *Client) forward_client_requests(requests chan interface{}) {
	for {
		var req Envelope
		client.conn.ReadJSON(&req)
		unwrappedReq := unwrapEnvelope(req)
		requests <- unwrappedReq
	}
}

func (clients *Client_list) create_client(conn *websocket.Conn, user string) {
	client := &Client{
		conn:  conn,
		user:  user,
		close: make(chan struct{}),
	}
	clients.users[user] = client
	go client.forward_client_requests(clients.requests)
}

func (clients *Client_list) handle_message(msg ChatMsg) {
	receiver, ok := clients.users[msg.body]
	if ok {
		receiver.conn.WriteJSON(msg)
	} else {
		// TODO
	}
}

func (clients *Client_list) handleConnReq(req ConnReq) {
	user := clients.users[req.from]
	reciever, ok := clients.users[req.to]

	if ok {
		reciever.conn.WriteJSON(req)
		res, ok := <-user.connRes
		if !ok {
			res.ok = false
			res.err = USER_DISCONNECTED
		}
		user.conn.WriteJSON(res)
	} else {
		res := ConnRes{
			ok:  false,
			err: USER_DISCONNECTED,
		}
		user.conn.WriteJSON(res)
	}
}

func (clients *Client_list) handleConnRes(res ConnRes) {
	receiver, ok := clients.users[res.to]

	if ok {
		receiver.connRes <- res
	}
	// If reciever disconnected, all connections ended automatically - no need to notify
}

func (clients *Client_list) handleDisconnReq(req DisconnReq) {
	// TODO
}

func (clients *Client_list) serve_clients() {
	for {
		request := <-clients.requests
		switch request.(type) {
		case ChatMsg:
			message := request.(ChatMsg)
			clients.handle_message(message)
		case ConnReq:
			req := request.(ConnReq)
			clients.handleConnReq(req)
		case ConnRes:
			res := request.(ConnRes)
			clients.handleConnRes(res)
		case DisconnReq:
			req := request.(DisconnReq)
			clients.handleDisconnReq(req)
		default:
			panic("Wrong requests type in requests channel")
		}
	}

}
