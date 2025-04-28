package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

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

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(message))
		resp := Response{Id: "12345"}
		err = conn.WriteJSON(resp)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var msg Request
	json.Unmarshal(body, &msg)

	response := Response{
		Id: fmt.Sprintf("recieved: %s", msg.Name),
	}
	response_str, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response_str)
	if err != nil {
		fmt.Println(err)
		return
	}
}

type Request struct {
	Name string
}

type Response struct {
	Id string
}

func main() {
	// scanner := bufio.NewScanner(os.Stdin)
	// port := get_port(scanner)

	router := http.NewServeMux()
	router.HandleFunc("/", handler)
	router.HandleFunc("/ws", serveWs)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Could not set up server on port. Quitting")
	}
}
