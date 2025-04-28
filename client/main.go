package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Request struct {
	Name string
}

type Response struct {
	Id string
}

type Message struct {
	sender_user   string
	receiver_user string
	message       string
}

const server_addr = "localhost:8080"

func make_request(req Request) *Response {
	client := &http.Client{}
	req_body, _ := json.Marshal(req)
	reader := bytes.NewReader(req_body)
	http_req, err := http.NewRequest("GET", "http://localhost:8080/", reader)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	resp, err := client.Do(http_req)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	resp_body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var response Response
	json.Unmarshal(resp_body, &response)
	return &response
}

func upgrade_conn() *websocket.Conn {
	ws_url := url.URL{Scheme: "ws", Host: server_addr, Path: "/ws"}
	dialer := websocket.DefaultDialer
	header := http.Header{}
	conn, _, err := dialer.Dial(ws_url.String(), header)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return conn
}

// Unnessesary for now
func forward_requests(conn *websocket.Conn, requests chan interface{}) {
	for {
		var request interface{}
		conn.ReadJSON(&request)
		requests <- request
	}
}

func handle_input(conn *websocket.Conn, user string, close_ch chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	const quit_str = "!quit"
	for scanner.Scan() {
		input := scanner.Text()

		if strings.HasPrefix(input, quit_str) {
			close(close_ch)
			return
		}

		msg := Message{
			sender_user:   user,
			receiver_user: "", // TODO ADD
			message:       input,
		}
		conn.WriteJSON(msg)
	}
}

func handle_conn(conn *websocket.Conn) {
	var request interface{}
	conn.ReadJSON(request)

	requests := make(chan interface{})
	close_ch := make(chan struct{})
	go forward_requests(conn, requests)
	go handle_input(conn, close_ch)

	select {
	case msg := <-requests:
		switch msg.(type) {
		case Message:
			msg := request.(Message)
			fmt.Println(msg.sender_user, ":")
			fmt.Println(msg.message)
		default:
			panic("Wrong requests type in requests channel")
		}
	case _, ok := <-close_ch:
		if !ok {
			panic("Incorrect usage of close channel")
		}

		// TODO reprompt user
		return
	}
}

func main() {
	conn := upgrade_conn()
	if conn == nil {
		fmt.Println("Connection failed")
		return
	}
	fmt.Println("Connection successful")
	handle_conn(conn)
	conn.WriteJSON(Request{Name: "we did it!"})
	resp := Response{}
	conn.ReadJSON(&resp)
	fmt.Println(resp)
}
