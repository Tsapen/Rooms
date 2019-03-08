package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

var listenAddr = "127.0.0.1:8082"

func main() {
	ctx, _ := context.WithCancel(context.Background())
	num := 0
	fmt.Println("Enter num of rooms")
	fmt.Fscan(os.Stdin, &num)
	err := StartServer(ctx, listenAddr, num)
	if err != nil {
		return
	}
	time.Sleep(1 * time.Hour)
}
