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

type Client struct {
	RPC      RSClient
	Name     string
	Password string
	Rooms    []string
}

func getCtx(roominfo bool, uName, rName, pswd string) context.Context {
	ctx := context.Background()
	md := metadata.Pairs()
	md = metadata.Pairs(
		"--username", uName,
		"--password", pswd,
	)
	if roominfo {
		md.Append("--roominfo", "")
	} else {
		md.Append("--roomname", rName)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (client *Client) SubscribeSession() {
	room := ""
	for {
		if message, err := client.RPC.Subscribe(getCtx(true, client.Name, "", client.Password), &Nothing{}); err != nil {
			fmt.Printf("Finish subscribe session: %v\n", err)
			return
		} else {
			fmt.Printf(message.Mess)
		}
		fmt.Println("Enter room name or --end for finish Subscribe session")
		fmt.Fscan(os.Stdin, &room)
		if room != "--end" {
			if _, err := client.RPC.Subscribe(getCtx(false, client.Name, room, client.Password), &Nothing{}); err != nil {
				fmt.Printf("Finish subscribe session: %v\n", err)
				return
			}
		} else {
			return
		}
	}
}

func (client *Client) PublishSession() {
	ctx := context.Background()
	errs := map[int]string{1: "room num isnt correct", 2: "room doesnt exist", 3: "user doesnt added"}
	publClient, err := client.RPC.Publish(ctx)
	if err != nil {
		fmt.Println("Other user with same name exist in room : ", err)
		return
	}
	room := ""
	fmt.Println("Enter room's name for publishing or --end for finish Publish session")
	_, flag := fmt.Fscan(os.Stdin, &room)
	if flag != nil || room == "--end" {
		fmt.Printf("___________\nFinish publish session\n")
		return
	}
	if err := publClient.Send(&Mess{room + "/" + client.Name + "/" + client.Password}); err != nil {
		fmt.Printf("Connection with room error:%v\n", err)
		fmt.Println("Finish publish session")
		return
	}
	mes, _ := publClient.Recv()
	if num, err := strconv.Atoi(mes.Mess); err != nil || num > 1 {
		fmt.Println("Connection with room error: " + errs[num])
		fmt.Println("Exit from publish session")
		return
	}
	fmt.Println("Chat:")
	fmt.Println("___________")
	// Getting history
	for {
		mes, err := publClient.Recv()
		if err != nil || mes.Mess == "--end" {
			break
		}
		fmt.Println(mes.Mess)
	}
	// Sending messages
	for {
		message := ""
		fmt.Fscan(os.Stdin, &message)
		err := publClient.Send(&Mess{message})
		if err != nil {
			fmt.Println("Publish error: ", err)
			break
		}
		if message == "--end" {
			fmt.Println("Publish session end")
			break
		}
	}

}

func NewClient(conn *grpc.ClientConn) *Client {
	client := &Client{
		RPC: NewRSClient(conn),
	}
	fmt.Printf("New session\nEnter your name:\n")
	fmt.Fscan(os.Stdin, &client.Name)
	fmt.Println("Password:")
	fmt.Fscan(os.Stdin, &client.Password)
	return client
}

func main() {
	listenAddr := "127.0.0.1:8082"
	conn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)
	defer conn.Close()
	if err != nil {
		log.Fatalf("main: Fail to dial: %v", err)
	}
	fmt.Println("You are connected to the server")
	choice := ""
LOOP:
	for {
		client := NewClient(conn)
		actions := map[string]func(){
			"1": client.SubscribeSession,
			"2": client.PublishSession,
		}
		for {
			fmt.Printf("===========\nChoose action:\n\"1\": Subscribe to new room\n\"2\": Publish something into one of your rooms\n\"3\": Change user\n\"4\": Exit\n")
			fmt.Fscan(os.Stdin, &choice)
			fmt.Println("===========")
			if choosenSession, ok := actions[choice]; ok {
				choosenSession()
				continue
			}
			if choice == "3" {
				continue LOOP
			}
			if choice == "4" {
				return
			}
			fmt.Println("Uncorrect choice")
		}
	}
}
