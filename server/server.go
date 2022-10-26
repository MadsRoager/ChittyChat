package main

import (
	proto "awesomeProject/proto"
	"context"
	"flag"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedMessagingServiceServer
	name      string
	port      int
	timestamp int
}

type Message struct {
	id        int64
	name      string
	message   string
	timestamp int
}

var channels [10]chan Message
var count = 0
var m sync.Mutex

var port = flag.Int("port", 8080, "server port number")

func main() {
	flag.Parse()

	server := &Server{
		name:      "server1",
		port:      *port,
		timestamp: 0,
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

	log.Printf("Lamport timestamp: %d, Server started", server.timestamp)

	proto.RegisterMessagingServiceServer(grpcServer, server)
	serverError := grpcServer.Serve(lister)

	if serverError != nil {
		log.Fatalln("could not start server")
	}
}

func (s *Server) SendMessage(ctx context.Context, in *proto.ClientSendMessage) (*proto.Ack, error) {
	s.updateTimestamp(int(in.TimeStamp), &m)
	var message = &Message{
		id:      in.Id,
		name:    in.ClientName,
		message: in.Message,
	}
	if in.Message == "leave" {
		message.message = in.ClientName + " left Chitty-Chat"
	} else {
		log.Printf("Lamport timestamp: %d, Client with name %s sent this message: %s\n", s.timestamp, in.ClientName, in.Message)
	}

	sendMessagesToChannels(message)

	return &proto.Ack{
		Success: "Succes!",
	}, nil
}

func sendMessagesToChannels(message *Message) {
	for i := 0; i < count; i++ {
		channels[i] <- *message
	}
}

func (s *Server) JoinChat(in *proto.ClientSendMessage, stream proto.MessagingService_JoinChatServer) error {
	sendMessagesToChannels(&Message{
		id:      in.Id,
		name:    in.ClientName,
		message: in.ClientName + " joined Chitty-Chat",
	})
	var index = creatingNewChannelAtIndex()
	for {
		var messageToBeBroadcasted = <-channels[index]

		if err := sendMessage(messageToBeBroadcasted, stream, s); err != nil {
			return err
		}
	}
}

func sendMessage(message Message, stream proto.MessagingService_JoinChatServer, s *Server) error {
	s.updateTimestamp(s.timestamp, &m)

	mes := &proto.ClientSendMessage{
		Id:         message.id,
		ClientName: message.name,
		Message:    message.message,
		TimeStamp:  int64(s.timestamp),
	}

	log.Printf("Lamport timestamp %d, Sending message", s.timestamp)

	if err := stream.Send(mes); err != nil {
		return err
	}
	return nil
}

func creatingNewChannelAtIndex() int {
	channel := make(chan Message)
	var index = count
	channels[index] = channel
	count++
	return index
}

func (s *Server) updateTimestamp(newTimestamp int, m *sync.Mutex) {
	m.Lock()
	s.timestamp = syncTimestamp(s.timestamp, newTimestamp)
	s.timestamp++
	m.Unlock()
}

func syncTimestamp(new int, old int) int {
	if new < old {
		return old
	}
	return new
}
