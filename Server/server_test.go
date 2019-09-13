package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func wait(amout int) {
	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
}

func getGrpcConn(t *testing.T) *grpc.ClientConn {
	grcpConn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("cant connect to grpc: %v", err)
	}
	return grcpConn
}

func getCtx(rName, uName, pswd string) context.Context {
	ctx := context.Background()
	md := metadata.Pairs(
		"--roomname", rName,
		"--username", uName,
		"--password", pswd,
	)
	return metadata.NewOutgoingContext(ctx, md)
}

func TestServerStartStop(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())
	err := StartServer(ctx, listenAddr)
	if err != nil {
		t.Fatalf("cant start server again: %v", err)
	}
	wait(1)
	finish()
	wait(1)
}

func TestSubscribe(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())
	err := StartServer(ctx, listenAddr)
	wait(1)
	if err != nil {
		t.Fatalf("cant start server initial: %v", err)
	}
	defer finish()
	defer wait(100)
	conn := getGrpcConn(t)
	client := NewRSClient(conn)
	ChErr := make(chan error, 2)
	go func() {
		_, err := client.Subscribe(getCtx("First", "Dale Cooper", "Good coffee!"), &Nothing{})
		ChErr <- err
	}()
	go func() {
		_, err := client.Subscribe(getCtx("Second", "Giant", "Owls are not what they seem"), &Nothing{})
		ChErr <- err
	}()
	for i := 0; i < 2; i++ {
		if err := <-ChErr; err != nil {
			t.Fatalf("Subscribe error: %v", err)
		}
	}
	_, err = client.Subscribe(getCtx("3rd", "Bob", "How is Annie?"), &Nothing{})
	if err == nil {
		t.Fatalf("Unexist room connect error: %v", err)
	}
	_, err = client.Subscribe(getCtx("First", "Dale Cooper", "Good coffee!"), &Nothing{})
	if err == nil {
		t.Fatalf("Second adding error not detected: %v", err)
	}
	_, err = client.Subscribe(getCtx("First", "Dale Cooper", "Excellent pie!"), &Nothing{})
	if err == nil {
		t.Fatalf("Simular name adding not detected: %v", err)
	}
}

func TestPublish(t *testing.T) {
	ctx, finish := context.WithCancel(context.Background())
	err := StartServer(ctx, listenAddr)
	wait(1)
	defer wait(100)
	defer finish()
	if err != nil {
		t.Fatalf("Cant start server initial: %v", err)
	}
	wait(1)
	conn := getGrpcConn(t)
	client := NewRSClient(conn)
	rooms, err := getConfig()
	if err != nil {
		t.Fatalf("Config file error: %v", err)
	}
	ChErr := make(chan error, len(rooms)-1)
	for userdata := 1; userdata < len(rooms); userdata++ {
		go func(userdata string, ChErr chan error) {
			for room := 1; room < len(rooms); room++ {
				ctx := getCtx(rooms[room], userdata, userdata)
				_, err = client.Subscribe(ctx, &Nothing{})
				if err != nil {
					ChErr <- err
					return
				}
				publ, err := client.Publish(ctx)
				if err != nil {
					ChErr <- err
					return
				}
				publ.Send(&Mess{rooms[room] + "/" + userdata + "/" + userdata})
				if err := publ.Send(&Mess{userdata}); err != nil {
					ChErr <- err
					return
				}
				publ.Send(&Mess{"--end"})
			}
			ChErr <- nil
			return
		}(strconv.Itoa(userdata), ChErr)
	}
	for i := 1; i < len(rooms); i++ {
		if err = <-ChErr; err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
	for i := 1; i < len(rooms); i++ {
		ctx := getCtx(rooms[i], strconv.Itoa(i), strconv.Itoa(i))
		publ, err := client.Publish(ctx)
		cnt := 0
		publ.Send(&Mess{rooms[i] + "/" + strconv.Itoa(i) + "/" + strconv.Itoa(i)})
		for {
			mes, err := publ.Recv()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if mes.Mess == "--end" {
				break
			}
			cnt++
		}
		publ.Send(&Mess{"--end"})
		cnt--
		if err != nil {
			t.Fatalf("Unexpectd error: %v", err)
		}
		if cnt != len(rooms)-1 {
			t.Fatalf("Num of messages(%d) less that expected(%d)", cnt, len(rooms)-1)
		}
	}
	wait(1)
	ctx = getCtx(rooms[0], strconv.Itoa(0), strconv.Itoa(0))
	client.Subscribe(ctx, &Nothing{})
	publ, err := client.Publish(ctx)
	if err != nil {
		t.Fatalf("New Publish client creating error: %v", err)
	}
	publ.Send(&Mess{rooms[0] + "/0/0"})
	x := ""
	for i := 0; i < 300; i++ {
		x += "a"
	}
	publ.Recv()
	publ.Recv()
	publ.Send(&Mess{x})
	publ.Send(&Mess{"--end"})
	publ, err = client.Publish(ctx)
	publ.Send(&Mess{rooms[0] + "/0/0"})
	num := 0
	for {
		num++
		if mess, err := publ.Recv(); err != nil {
			t.Fatalf("Error: %v", err)
		} else if mess.Mess == "--end" {
			break
		}
	}
	num -= 2
	if num > 0 {
		t.Fatalf("Adding in query more than 254 byte message")
	}
	for i := 0; i < 150; i++ {
		if publ.Send(&Mess{strconv.Itoa(i)}) != nil {
			t.Fatalf("Error sending")
		}
		wait(1)
	}
	publ.Send(&Mess{"--end"})
	publ, err = client.Publish(ctx)
	publ.Send(&Mess{rooms[0] + "/0/0"})
	last := ""
	for {
		mess, _ := publ.Recv()
		if mess.Mess == "--end" {
			break
		}
		num++
		last = mess.Mess
	}
	num--
	if num != 127 || last != "0: 149" {
		t.Fatalf("Uncorrect publishing: expected last message \"%s\", received \"%s\"", "0: 149", last)
	}
}
