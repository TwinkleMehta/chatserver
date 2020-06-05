package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"google.golang.org/grpc"
	proto "github.com/TwinkleMehta/chatserver/proto"
)

var client proto.BroadcastClient 
var wait *sync.WaitGroup 

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *proto.User) error{
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &proto.Connect{
		User: user,
		Active: true, 
	})

	if err != nil{
		return fmt.Errorf("connection failed: %v",err)
	}

	wait.Add(1)
	go func(str proto.Broadcast_CreateStreamClient){
		defer wait.Done()
		for{
			msg, err := str.Recv()
			if err != nil{
				streamerror = fmt.Errorf("Error reading message: %v", err)
				break
			}

			fmt.Printf("(%v) %v : %s\n", msg.Timestamp, msg.Id, msg.Content)
		}
	}(stream)

	return streamerror
}

func main(){
	timestamp := time.Now().Format("Mon Jan 2 3:04 PM")
	done := make(chan int)

	name := flag.String("N", "Anon", "The name of the user")
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp + *name))

	conn,err := grpc.Dial("34.66.13.11:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = proto.NewBroadcastClient(conn)
	user := &proto.User{
		Id: hex.EncodeToString(id[:]),
		Name: *name,
	}

	connect(user)

	wait.Add(1)
	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &proto.Message{
				Id: user.Name, // User's name is Message ID
				Content: scanner.Text(),
				Timestamp: timestamp,
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil{
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}