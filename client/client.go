package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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

	go listenForBroadcasts()

	parseInput()
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

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		sendChatMessage(input)
	}
}

func sendChatMessage(message string) {
	if message == "exit" {
		server.Leave(context.Background(), &gRPC.LeaveRequest{Name: *clientsName})
		os.Exit(0)
	}

	stream, err := server.Broadcast(context.Background())
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}

	msg := &gRPC.ChatMessage{Name: *clientsName, Message: message}
	stream.Send(msg)
}

func listenForBroadcasts() {
	stream, err := server.Broadcast(context.Background())
	if err != nil {
		log.Println("Failed to get broadcast stream:", err)
		return
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Failed to receive broadcast:", err)
			return
		}

		fmt.Printf("[Lamport Timestamp: %d] %s\n", msg.Timestamp, msg.Message)
	}
}
