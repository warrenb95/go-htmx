package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var (
	messages = []Message{}
)

type Message struct {
	Message string
}

func index(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseGlob("static/templates/*")
	if err != nil {
		log.Printf("failed to parse glob templates: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("executing index")
	err = tmp.ExecuteTemplate(w, "index", nil)
	if err != nil {
		log.Fatalf("failed to execute index template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func chatroom(ws *websocket.Conn) {
	log.Println("executing chatroom")

	var message Message
	err := websocket.JSON.Receive(ws, &message)
	if err != nil {
		log.Printf("failed to receive message from websocket: %v", err)
		return
	}

	messages = append(messages, message)

	log.Printf("chatroom message: %s", message.Message)
	log.Printf("chatroom message length: %d", len(messages))

	tmpl, err := template.ParseFiles("static/templates/chatroom.html")
	if err != nil {
		log.Printf("failed to parse chatroom template: %v", err)
		return
	}
	log.Println("parsed chatroom template")

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "chat_room", messages)
	if err != nil {
		log.Printf("failed to execute template: %v", err)
		return
	}

	err = websocket.Message.Send(ws, buf.String())
	if err != nil {
		log.Printf("failed to send websocket response: %v", err)
		return
	}

}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/chatroom", websocket.Handler(chatroom))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
