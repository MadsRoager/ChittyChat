package main

import (
	proto "awesomeProject/proto"
	"context"
	"flag"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedMessagingServiceServer
	name string
	port int
}

type Message struct {
	id      int64
	name    string
	message string
}

var channels [10]chan Message
var count = 0

var port = flag.Int("port", 8080, "server port number")

func main() {
	flag.Parse()

	server := &Server{
		name: "server1",
		port: *port,
	}

	go startServer(server)

	for {
		time.Sleep(100 * time.Second)
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
		var mes = &Message{
			id:      in.Id,
			name:    in.ClientName,
			message: in.Message,
		}
		for i := 0; i < count; i++ {
			log.Printf("sending message in channel, index is %d, count is %d", i, count)
			channels[i] <- *mes
			log.Println("has sent to channel")
		}
	}

	return &proto.Ack{
		Success: "Succes!",
	}, nil
}

func (s *Server) JoinChat(in *proto.ClientSendMessage, stream proto.MessagingService_JoinChatServer) error {
	channel1 := make(chan Message)
	var index = count
	channels[index] = channel1
	count++
	log.Printf("channel array has length %d", len(channels))
	log.Printf("count is  %d", count)
	for {
		// wait to rec
		var chanMes = <-channels[index]

		mes := &proto.ClientSendMessage{
			Id:         chanMes.id,
			ClientName: chanMes.name,
			Message:    chanMes.message,
		}
		if err := stream.Send(mes); err != nil {
			return err
		}

	}

}
