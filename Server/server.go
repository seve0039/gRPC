package main

import (

	// This has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	"flag"
	"fmt"
	"io"
	"net"
	"sync"

	gRPC "github.com/seve0039/gRPC/proto"

	"google.golang.org/grpc"
)
type Server struct {
    // an interface that the server type needs to have
    gRPC.UnimplementedTemplateServer
    name                             string // Not required but useful if you want to name your server
	port                             string // Not required but useful if your server needs to know what port it's listening to

	incrementValue int64      // value that clients can increment.
	mutex          sync.Mutex // used to lock the server to avoid race conditions.
    // here you can implement other fields that you want
}
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal

func main() {

	flag.Parse()
	fmt.Println(".:server is starting:.")

    list, _ := net.Listen("tcp", "localhost:5400")
    var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	server := &Server{
			name:           *serverName,
			port:           *port,
			incrementValue: 0,
	}
	gRPC.RegisterTemplateServer(grpcServer, server)

	grpcServer.Serve(list)
	// Code here will not run as .Serve() blocks the thread

}
func (s *Server) SayHi(msgStream gRPC.Template_SayHiServer) error {
    // get all the messages from the stream
    for {
        msg, err := msgStream.Recv()
        if err == io.EOF {
            break
        }
    }

    ack := &gRPC.Farewell{Message: "Goodbye"}
    msgStream.SendAndClose(ack)

    return nil
}




