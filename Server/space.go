package main

import (
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"context"
	"net"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type Room struct {
	mes   []Msg
	first int
	last  int
	users map[string]string
}

type Msg struct {
	author string
	info   string
}

type RS struct {
	space map[string]*Room
	mu    *sync.Mutex
}

func getConfig() ([]string, error) {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		return []string{}, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	config := make([]string, 10)
	if err = json.Unmarshal(byteValue, &config); err != nil {
		return []string{}, err
	}
	return config, err
}

func StartServer(ctx context.Context, Addr string) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	rs := RS{
		space: map[string]*Room{},
		mu:    &sync.Mutex{},
	}
	for _, i := range config {
		rs.space[i] = &Room{
			mes:   make([]Msg, 128, 128),
			users: map[string]string{},
		}
	}
	server := grpc.NewServer()
	lis, err := net.Listen("tcp", Addr)
	if err != nil {
		return err
	}
	go func() {
		RegisterRSServer(server, &rs)
		fmt.Println("Starting server at :8081")
		server.Serve(lis)
	}()
	go func() {
		select {
		case <-ctx.Done():
			server.Stop()
		}
	}()
	return nil
}

func (rs RS) RoomInfo(username, password string) string {
	sub, avail := "", ""
	rs.mu.Lock()
	for roomname, i := range rs.space {
		if pswd, ok := i.users[username]; ok && pswd == password {
			sub += roomname + ", "
		} else {
			avail += roomname + ", "
		}
	}
	rs.mu.Unlock()
	if sub == "" {
		sub = "You haven't subscriptions"
	} else {
		sub = "Current subscriptions: " + sub[:len(sub)-2]
	}
	if avail == "" {
		avail = "There are not rooms for subscribing"
	} else {
		avail = "Can subscribe: " + avail[:len(avail)-2]
	}
	return sub + "\n" + avail + "\n"
}

func (rs RS) Subscribe(ctx context.Context, n *Nothing) (*Mess, error) {
	MD, _ := metadata.FromIncomingContext(ctx)
	buf, _ := MD["--username"]
	user := buf[0]
	buf, _ = MD["--password"]
	pswd := buf[0]
	if _, ok := MD["--roominfo"]; ok {
		return &Mess{rs.RoomInfo(user, pswd)}, nil
	}
	buf, _ = MD["--roomname"]
	room := buf[0]
	fmt.Println("Subscriber: " + room + ", " + user + ", " + pswd)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	_, ok := rs.space[room]
	if !ok {
		return nil, grpc.Errorf(codes.Unknown, "Room doesnt exist")
	}
	if _, ok := rs.space[room].users[user]; ok {
		return nil, grpc.Errorf(codes.AlreadyExists, "User already added")
	}
	rs.space[room].users[user] = pswd
	return &Mess{}, nil
}

func (rs RS) Publish(stream RS_PublishServer) error {
	val, _ := stream.Recv()
	pars := strings.Split(val.Mess, "/")
	room := pars[0]
	user := pars[1]
	pswd := pars[2]
	fmt.Println("Publisher:", pars)
	rs.mu.Lock()
	_, ok := rs.space[room]
	rs.mu.Unlock()
	if !ok {
		stream.Send(&Mess{strconv.Itoa(2)})
		return grpc.Errorf(codes.Unauthenticated, "Room doesnt exist")
	}
	rs.mu.Lock()
	chpswd, ok := rs.space[room].users[user]
	rs.mu.Unlock()
	if !ok || chpswd != pswd {
		stream.Send(&Mess{strconv.Itoa(3)})
		return grpc.Errorf(codes.Unauthenticated, "User doesnt added")
	}
	stream.Send(&Mess{strconv.Itoa(0)})
	rs.mu.Lock()
	f := rs.space[room].first
	l := rs.space[room].last
	for l != f {
		txt := Mess{rs.space[room].mes[f].author + ": " + rs.space[room].mes[f].info}
		stream.Send(&txt)
		f = (f + 1) % 128
	}
	rs.mu.Unlock()
	stream.Send(&Mess{"--end"})
	for {
		mes, err := stream.Recv()
		if err != nil {
			return err
		}
		if mes.Mess == "--end" {
			return nil
		}
		if len(mes.Mess) > 254 {
			continue
		}
		rs.mu.Lock()
		rs.space[room].mes[rs.space[room].last].author = user
		rs.space[room].mes[rs.space[room].last].info = mes.Mess
		rs.space[room].last = (rs.space[room].last + 1) % 128
		if rs.space[room].last == rs.space[room].first {
			rs.space[room].first = (rs.space[room].first + 1) % 128
		}
		rs.mu.Unlock()
	}

}
