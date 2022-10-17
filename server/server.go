package main

import (
	proto "awesomeProject/proto"
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"time"
)

type Server struct {
	proto.UnimplementedMessagingServiceServer
	name string
	port int
}

var port = flag.Int("port", 8080, "server port number")

func main() {
	flag.Parse()

	server := &Server{
		name: "server1",
		port: *port,
	}

	go startServer(server)

	for {
		time.Sleep(100)
	}
}

func startServer(server *Server) {
	grpcServer := grpc.NewServer()

	lister, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalln("could not start listener")
	}

	log.Printf("Server started")

	proto.RegisterMessagingServiceServer(grpcServer, server)
	serverError := grpcServer.Serve(lister)

	if serverError != nil {
		log.Fatalln("could not start server")
	}
}

func (s *Server) SendMessage(ctx context.Context, in *proto.ClientSendMessage) (*proto.Ack, error) {
	switch in.Message {
	case "join":
	//tilføj til array
	case "leave":
	//fjern fra array
	default:
		log.Printf("Client with name %s sent this message: %s to the server\n", in.ClientName, in.Message)
	}

	return &proto.Ack{
		Success: "Succes!",
	}, nil
}
