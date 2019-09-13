package main

import (
	"context"
	"log"
)

var listenAddr = "127.0.0.1:8082"

func main() {
	ctx, _ := context.WithCancel(context.Background())
	err := StartServer(ctx, listenAddr)
	if err != nil {
		return
	}
	log.Printf("To stop server press CTRL+C")
	stop := make(chan bool)
	<-stop

}
