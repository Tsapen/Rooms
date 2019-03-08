package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func getCtx(rName, uName, pswd string) context.Context {
	ctx := context.Background()
	md := metadata.Pairs(
		"roomname", rName,
		"username", uName,
		"password", pswd,
	)
	return metadata.NewOutgoingContext(ctx, md)
}

func main() {
	ctx, _ := context.WithCancel(context.Background())
	listenAddr := "127.0.0.1:8082"
	conn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)
	errs := map[int]string{1: "room num isnt correct", 2: "room doesnt exist", 3: "user doesnt added"}
	defer conn.Close()
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	fmt.Println("You are connected to the server")
LOOP:
	for {
		client := NewRSClient(conn)
		fmt.Println("Enter your name...")
		user, pswd, room := "", "", ""
		fmt.Fscan(os.Stdin, &user)
		fmt.Println("... and password")
		fmt.Fscan(os.Stdin, &pswd)
		fmt.Printf("Hello, %s. Choose zero or more room nums for Subscribing. After this enter --next)\n", user)
	LOOP0:
		for {
			fmt.Fscan(os.Stdin, &room)
			if room != "--next" {
				_, err := client.Subscribe(getCtx(room, user, pswd), &Nothing{})
				if err != nil {
					fmt.Printf("Stop client session: %v\n", err)
					continue LOOP
				}
			} else {
				break LOOP0
			}
		}
		publ, err := client.Publish(ctx)
		if err != nil {
			fmt.Println("Other user with same name exist in room : ", err)
		} else {

			fmt.Println("Enter room for Publishing")
			fmt.Fscan(os.Stdin, &room)
			if err := publ.Send(&Mess{room + "/" + user + "/" + pswd}); err != nil {
				fmt.Printf("Connection with room error:%v\n", err)
				fmt.Println("Stop client session")
				continue LOOP
			}
			mes, _ := publ.Recv()
			if num, err := strconv.Atoi(mes.Mess); err != nil || num > 1 {
				fmt.Println("Connection with room error:" + errs[num])
				fmt.Println("Stop client session")
				continue LOOP
			}
			fmt.Println("History:")
			fmt.Println("___________")
		LOOP1:
			for {
				mes, err := publ.Recv()
				if err != nil || mes.Mess == "--empty" {
					fmt.Println("___________")
					break LOOP1
				}
				fmt.Println(mes.Mess)
			}
			fmt.Println("Ok, publish something (\"--end\" is stop-command)")
			fmt.Println("___________")
		LOOP2:
			for {
				txt := ""
				fmt.Fscan(os.Stdin, &txt)
				err := publ.Send(&Mess{txt})
				if err != nil {
					fmt.Println("Publish error: ", err)
					break LOOP2
				}
				if txt == "--end" {
					fmt.Println("Session end")
					break LOOP2
				}
			}
			fmt.Println("___________")
		}
		fmt.Println("Else? For exit --exit")
		buf := ""
		_, flag := fmt.Fscan(os.Stdin, &buf)
		if flag != nil || buf == "--exit" {
			fmt.Println("Goodbye")
			break
		}
	}
}
