package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/kittsville/PlugChan/commons"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func handleInput(inputChannel chan *commons.PlugEvent, w http.ResponseWriter, r *http.Request) {
	plug_number, err := urlIntParam(r, "plug")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	state, err := urlIntParam(r, "state")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if plug_number < 0 || plug_number > 4 {
		http.Error(w, fmt.Sprintf("Plug number must be 0-4 (inclusive), given %d", plug_number), 400)
		return
	}

	if state != 0 && state != 1 {
		http.Error(w, "Plug state must be 0 or 1", 400)
		return
	}

	inputChannel <- &commons.PlugEvent{
		Plug:  plug_number,
		State: state == 1,
	}

	fmt.Fprint(w, "Sent")
}

var upgrader = websocket.Upgrader{} // use default options

func handleOutput(inputChannel chan *commons.PlugEvent, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for event := range inputChannel {
		message := fmt.Sprintf("Received plug %d event to change state to %t\n", event.Plug, event.State)
		fmt.Print(message)
		eventBytes, err := json.Marshal(event)

		if err != nil {
			fmt.Println("Failed to marshal plug event JSON")
		}

		c.WriteMessage(websocket.TextMessage, eventBytes)
	}
}

func urlIntParam(r *http.Request, param_name string) (int, error) {
	param_values, ok := r.URL.Query()[param_name]

	if !ok || len(param_values[0]) < 1 {
		return -1, fmt.Errorf("Url Param '%s' is missing", param_name)
	}

	int, err := strconv.Atoi(param_values[0])

	if err != nil {
		return -1, fmt.Errorf("Url Param '%s' is not an integer", param_name)
	}

	return int, nil
}

func listener(inputChannel chan *commons.PlugEvent) {
	for event := range inputChannel {
		fmt.Printf("Received plug %d event to change state to %t\n", event.Plug, event.State)
	}
}

func main() {
	flag.Parse()
	inputChannel := make(chan *commons.PlugEvent)

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		handleInput(inputChannel, w, r)
	})
	http.HandleFunc("/output", func(w http.ResponseWriter, r *http.Request) {
		handleOutput(inputChannel, w, r)
	})
	log.Printf("Listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
