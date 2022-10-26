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

func (s *Server) sendMessagesToAllStreams(messageToBeBroadcasted *Message) {
	for i := 0; i < count; i++ {
		sendMessage(*messageToBeBroadcasted, streams[i], s)
	}
}

func (s *Server) Chat(stream proto.MessagingService_ChatServer) error {
	streams[count] = stream
	count++
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		s.updateTimestamp(int(in.TimeStamp), &m)
		if in.Message == "leave" {
			log.Printf("Lamport timestamp: %d, Client with name %s left the chat", s.timestamp, in.ClientName)
			s.sendMessagesToAllStreams(&Message{
				id:      in.Id,
				name:    in.ClientName,
				message: in.ClientName + " left the Chitty-chat",
			})

		} else {
			log.Printf("Lamport timestamp: %d, Client with name %s sent this message: %s\n", s.timestamp, in.ClientName, in.Message)
			s.sendMessagesToAllStreams(&Message{
				id:      in.Id,
				name:    in.ClientName,
				message: in.Message,
			})
		}
	}
}

func sendMessage(message Message, stream proto.MessagingService_ChatServer, s *Server) {
	s.updateTimestamp(s.timestamp, &m)

	mes := &proto.Message{
		Id:         message.id,
		ClientName: message.name,
		Message:    message.message,
		TimeStamp:  int64(s.timestamp),
	}
	// only sends message if stream is open
	select {
	case <-stream.Context().Done():
		return
	default:
		log.Printf("Lamport timestamp %d, Sending message", s.timestamp)

		stream.Send(mes)
	}

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
