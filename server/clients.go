package main

import "github.com/gorilla/websocket"

type Message struct {
	sender_user   string
	receiver_user string
	message       string
}

type Client struct {
	conn  websocket.Conn
	user  string
	close chan struct{}
}

type Client_list struct {
	users       map[string]Client
	connections map[string]string // Maps connected users to their connections
}

func (clients *Client_list) create_client(conn websocket.Conn, user string) {
	client := Client{
		conn:  conn,
		user:  user,
		close: make(chan struct{}),
	}
	clients.users[user] = client
}

func (client *Client) forward_client_requests(requests chan interface{}) {
	for {
		var request interface{}
		client.conn.ReadJSON(&request)
		requests <- request
	}
}

func (clients *Client_list) handle_message(message Message) {
	receiver, ok := clients.users[message.message]
	if ok {
		receiver.conn.WriteJSON(message)
	} else {
		// TODO
	}
}

func serve_clients() {
	clients := Client_list{}
	clients.users = make(map[string]Client)
	clients.connections = make(map[string]string)

	// Messages and connection requests
	requests := make(chan interface{})

	for _, client := range clients.users {
		go client.forward_client_requests(requests)
	}

	for {
		request := <-requests
		switch request.(type) {
		case Message:
			message := request.(Message)
			clients.handle_message(message)
		default:
			panic("Wrong requests type in requests channel")
		}
	}

}
