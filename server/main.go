package main

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Request struct {
	Name string
}

type Response struct {
	Id string
}

func get_port(scanner *bufio.Scanner) int {
	for scanner.Scan() {
		input := scanner.Text()
		var port int
		_, err := fmt.Sscanf(input, "%d", &port)
		if err != nil {
			fmt.Println("input must be a positive integer")
			continue
		}
		return port
	}

	panic("Should not reach end of standard input")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(clients *Client_list, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	id := r.PathValue("username")
	clients.create_client(conn, id)
}

func main() {
	// scanner := bufio.NewScanner(os.Stdin)
	// port := get_port(scanner)
	clients := initializeClients()
	clients.serve_clients()
	serveWsWrapper := func(w http.ResponseWriter, r *http.Request) {
		serveWs(clients, w, r)
	}

	router := http.NewServeMux()
	router.HandleFunc("GET /ws/{username}", serveWsWrapper)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Could not set up server on port. Quitting")
	}
}
