package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	gRPC "github.com/seve0039/gRPC.git/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittyChatServer
	participants     map[string]gRPC.ChittyChat_BroadcastServer
	participantMutex sync.RWMutex
	name             string
	port             string
	lamportClock     int64
}

var serverName = flag.String("name", "default", "Server's name")
var port = flag.String("port", "5400", "Server port")

func main() {
	createLogFile()
	flag.Parse()
	fmt.Println("--- server is starting ---")
	launchServer()
}

func createLogFile() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
}

func launchServer() {
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", *port, err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{
		name:         *serverName,
		port:         *port,
		participants: make(map[string]gRPC.ChittyChat_BroadcastServer),
	}

	gRPC.RegisterChittyChatServer(grpcServer, server)
	log.Printf("NEW SESSION: Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Join(ctx context.Context, joinReq *gRPC.JoinRequest) (*gRPC.JoinAck, error) {
	ack := &gRPC.JoinAck{Message: fmt.Sprintf("Welcome to Chitty-Chat, %s!", joinReq.Name)}
	s.broadcastMessage(fmt.Sprintf("Participant %s joined Chitty-Chat at Lamport time %d", joinReq.GetName(), s.incrementLamport()))
	return ack, nil
}

func (s *Server) Leave(ctx context.Context, leaveReq *gRPC.LeaveRequest) (*gRPC.LeaveAck, error) {
	
	
	delete(s.participants, leaveReq.Name)
	fmt.Println("Participant ", leaveReq.Name, " left Chitty-Chat")
	s.broadcastMessage(fmt.Sprintf("Participant %s left Chitty-Chat at Lamport time %d", leaveReq.GetName(), s.incrementLamport()))

	return &gRPC.LeaveAck{Message: "Goodbye!"}, nil

}

func (s *Server) Broadcast(stream gRPC.ChittyChat_BroadcastServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	s.participantMutex.Lock()
	s.participants[msg.Name] = stream
	s.participantMutex.Unlock()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		s.broadcastMessage(fmt.Sprintf("%s: %s at Lamport time %d", msg.Name, msg.Message, s.incrementLamport()))
	}
}

func (s *Server) broadcastMessage(message string) {
	s.participantMutex.RLock()
	defer s.participantMutex.RUnlock()

	fmt.Printf("Received: %v\n", message)
	log.Printf("Received: %v", message)

	for _, participant := range s.participants {
		participant.Send(&gRPC.ChatMessage{Message: message, Timestamp: s.lamportClock})
	}
}

func (s *Server) incrementLamport() int64 {
	s.lamportClock++
	return s.lamportClock
}
