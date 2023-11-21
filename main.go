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

	conns map[*websocket.Conn]struct{}
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

func chat(ws *websocket.Conn) {
	log.Println("connecting to chatroom")

	conns[ws] = struct{}{}

	for {
		// Get the message from the ws. Can't just 'read', must uses the JSON/Message.Receive function.
		var message Message
		err := websocket.JSON.Receive(ws, &message)
		if err != nil {
			log.Printf("failed to receive message from websocket: %v", err)
			return
		}

		messages = append(messages, message)

		tmpl, err := template.ParseFiles("static/templates/chatroom.html")
		if err != nil {
			log.Printf("failed to parse chatroom template: %v", err)
			return
		}
		log.Println("parsed chatroom template")

		// Need to write the whole template to the buffer first and then respond by sending the buf as a message to the
		// ws below.
		var buf bytes.Buffer
		err = tmpl.ExecuteTemplate(&buf, "chat_room", messages)
		if err != nil {
			log.Printf("failed to execute template: %v", err)
			return
		}

		for connection := range conns {
			err = websocket.Message.Send(connection, buf.String())
			if err != nil {
				log.Printf("failed to send websocket response: %v", err)
			}
		}
	}

}

func main() {
	conns = make(map[*websocket.Conn]struct{})

	http.HandleFunc("/", index)
	http.Handle("/chatroom", websocket.Handler(chat))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
