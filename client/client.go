package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	gRPC "github.com/seve0039/gRPC.git/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var user = generateRandomString(5)
var clientsName = flag.String("name", user, "Sender's name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChittyChatClient
var ServerConn *grpc.ClientConn

func main() {
	//fmt.Println("--- Please enter username")
	//readerName := bufio.NewReader(os.Stdin)
	flag.Parse()
	fmt.Println("--- LOGGED IN AS: ", user, " ----")

	ConnectToServer()
	defer ServerConn.Close()

	joinServer()

	stream, err := server.Broadcast(context.Background())
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}

	go listenForBroadcasts(stream)
	sendChatMessage("A client joined", stream)
	parseInput(stream)
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

	if message == "exit" {
		server.Leave(context.Background(), &gRPC.LeaveRequest{Name: *clientsName})
		os.Exit(0)

	} else {
		msg := &gRPC.ChatMessage{Name: *clientsName, Message: message}
		stream.Send(msg)
	}

}

func listenForBroadcasts(stream gRPC.ChittyChat_BroadcastClient) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Failed to receive broadcast: ", err)
			return
		}

		fmt.Println(msg.GetMessage())
	}
}

// This function was copied from here https://www.tutorialspoint.com/how-to-generate-random-string-characters-in-golang
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
