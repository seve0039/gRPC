package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	gRPC "github.com/seve0039/gRPC.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var clientsName = flag.String("name", "default", "Sender's name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChittyChatClient
var ServerConn *grpc.ClientConn

func main() {
	flag.Parse()
	fmt.Println("--- CLIENT APP ---")

	ConnectToServer()
	defer ServerConn.Close()

	joinServer()

	stream, err := server.Broadcast(context.Background())
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}

	go listenForBroadcasts(stream)

	parseInput(stream)

	log.Println("33")
	fmt.Println("34")
}

func ConnectToServer() {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Fatalf("Fail to Dial : %v", err)
	}

	server = gRPC.NewChittyChatClient(conn)
	ServerConn = conn
}

func joinServer() {
	_, err := server.Join(context.Background(), &gRPC.JoinRequest{Name: *clientsName})
	if err != nil {
		log.Fatalf("Failed to join server: %v", err)
	}
}

func parseInput(stream gRPC.ChittyChat_BroadcastClient) {
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		sendChatMessage(reader.Text(), stream)
	}
}

func sendChatMessage(message string, stream gRPC.ChittyChat_BroadcastClient) {
	/*
		if message == "exit" {
			server.Leave(context.Background(), &gRPC.LeaveRequest{Name: *clientsName})
			os.Exit(0)
		}
	*/

	msg := &gRPC.ChatMessage{Name: *clientsName, Message: message}
	log.Println("Ready for sendoff")
	stream.Send(msg)
	log.Println("Message has been sent")
}

func listenForBroadcasts(stream gRPC.ChittyChat_BroadcastClient) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Failed to receive broadcast:", err)
			return
		}

		log.Printf("[Lamport Timestamp: %d] %s\n", msg.GetTimestamp(), msg.GetMessage())
	}
}
