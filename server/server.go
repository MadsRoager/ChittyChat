package main

import (
	proto "awesomeProject/proto"
	"flag"
	"io"
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
	id      int64
	name    string
	message string
}

var streams [100]proto.MessagingService_ChatServer
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

func (server *Server) Chat(stream proto.MessagingService_ChatServer) error {
	streams[count] = stream
	count++
	for {
		messageFromClient, err := receiveMessageFromClient(stream)
		if err != nil {
			return err
		}

		server.updateTimestamp(int(messageFromClient.TimeStamp), &m)

		server.handleMessage(messageFromClient)
	}
}

func receiveMessageFromClient(stream proto.MessagingService_ChatServer) (*proto.Message, error) {
	messageFromClient, err := stream.Recv()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return messageFromClient, nil
}

func (server *Server) handleMessage(messageFromClient *proto.Message) {
	var messageToBeBroadcasted string
	if messageFromClient.Message == "leave" {
		log.Printf("Lamport timestamp: %d, Client with name %s left the chat",
			server.timestamp, messageFromClient.ClientName)
		messageToBeBroadcasted = messageFromClient.ClientName + " left the Chitty-chat"
	} else {
		log.Printf("Lamport timestamp: %d, Client with name %s sent this message: %s\n",
			server.timestamp, messageFromClient.ClientName, messageFromClient.Message)
		messageToBeBroadcasted = messageFromClient.Message
	}

	server.sendMessagesToAllStreams(&Message{
		id:      messageFromClient.Id,
		name:    messageFromClient.ClientName,
		message: messageToBeBroadcasted,
	})
}

func (server *Server) sendMessagesToAllStreams(messageToBeBroadcasted *Message) {
	for i := 0; i < count; i++ {
		sendMessage(*messageToBeBroadcasted, streams[i], server)
	}
}

func sendMessage(message Message, stream proto.MessagingService_ChatServer, server *Server) {
	// only sends message if stream is open
	select {
	case <-stream.Context().Done():
		return
	default:
		server.updateTimestamp(server.timestamp, &m)

		log.Printf("Lamport timestamp %d, Sending message", server.timestamp)

		stream.Send(&proto.Message{
			Id:         message.id,
			ClientName: message.name,
			Message:    message.message,
			TimeStamp:  int64(server.timestamp),
		})
	}

}

func (server *Server) updateTimestamp(newTimestamp int, m *sync.Mutex) {
	m.Lock()
	server.timestamp = maxValue(server.timestamp, newTimestamp)
	server.timestamp++
	m.Unlock()
}

func maxValue(new int, old int) int {
	if new < old {
		return old
	}
	return new
}
