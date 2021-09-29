package main

import (
	"asd/client_http"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type msg struct {
	Symbol string
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", rootHandler)

	panic(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	done := make(chan bool)
	go echo(conn, done)

	<-r.Context().Done()
	close(done)
}

func echo(conn *websocket.Conn, done chan bool) {
	m := msg{}
	ticker := time.NewTicker(3 * time.Second)
	data := make(chan *client_http.OrderBook)

	err := conn.ReadJSON(&m)
	if err != nil {
		fmt.Println("Error reading json.", err)
	}

	fmt.Printf("Got message: %#v\n", m)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)

				orderBooks, err := client_http.MakeRequest(m.Symbol)
				if err != nil {
					fmt.Println(err)
					continue
				}
				// положить в канал orderBooks
				data <- orderBooks

				// отображение на консоли
				fmt.Printf("Bids: %v\n\nAsks: %v\n\nSum Bids Quantity: %f\n\nSum Asks Quantity: %f\n",
					orderBooks.Bids, orderBooks.Asks, orderBooks.SumBidsQuantity, orderBooks.SumAsksQuantity)
			}
		}
	}()

outer:
	for {
		select {
		case <-done:
			break outer
		case d := <-data:
			if err := conn.WriteJSON(d); err != nil {
				fmt.Println(err)
				close(done)
			}
		}
	}
}

// ресурс https://gist.github.com/tmichel/7390690
