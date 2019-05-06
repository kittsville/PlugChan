package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

type plugEvent struct {
	plug  int
	state bool
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func handleInput(inputChannel chan *plugEvent, w http.ResponseWriter, r *http.Request) {
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

	inputChannel <- &plugEvent{
		plug:  plug_number,
		state: state == 1,
	}

	fmt.Fprint(w, "Sent")
}

var upgrader = websocket.Upgrader{} // use default options

func handleOutput(inputChannel chan *plugEvent, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for event := range inputChannel {
		message := fmt.Sprintf("Received plug %d event to change state to %t\n", event.plug, event.state)
		fmt.Print(message)
		c.WriteMessage(websocket.TextMessage, []byte(message))
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

func listener(inputChannel chan *plugEvent) {
	for event := range inputChannel {
		fmt.Printf("Received plug %d event to change state to %t\n", event.plug, event.state)
	}
}

func main() {
	flag.Parse()
	inputChannel := make(chan *plugEvent)

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		handleInput(inputChannel, w, r)
	})
	http.HandleFunc("/output", func(w http.ResponseWriter, r *http.Request) {
		handleOutput(inputChannel, w, r)
	})
	log.Printf("Listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
