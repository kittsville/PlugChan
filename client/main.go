package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/kittsville/PlugChan/commons"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/output"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var event commons.PlugEvent

		err = json.Unmarshal(message, &event)

		if err != nil {
			log.Println("Failed to unmarshal plug event JSON")
			continue
		}

		log.Printf("%+v\n", event)
	}
}
