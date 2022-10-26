package main

import (
	"awesomeProject/proto"
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name       string
	portNumber int
	timestamp  int
}

var (
	clientPort = flag.Int("clientPort", 8081, "client port number")
	serverPort = flag.Int("serverPort", 8080, "server port number")
	clientName = flag.String("clientName", "DefaultName", "client name")
)

var m sync.Mutex

func main() {

	flag.Parse()
	client := &Client{
		name:       *clientName,
		portNumber: *clientPort,
		timestamp:  0,
	}
	go startClient(client)
	for {
		time.Sleep(100 * time.Second)
	}
}

func startClient(client *Client) {

	log.Printf("Lamport timestamp %d, Client started", client.timestamp)

	serverConnection := getServerConnection(client)

	go establishConnectionToChat(client, serverConnection)
}

func readInput(client *Client, serverConnection proto.MessagingServiceClient, stream proto.MessagingService_ChatClient) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		client.updateTimestamp(client.timestamp, &m)

		input := scanner.Text()
		if input == "leave" {
			handleLeave(client, serverConnection, stream)
			time.Sleep(1 * time.Second)
			os.Exit(0)
		} else {
			handleMessageInput(client, serverConnection, input, stream)
		}
	}
}

func handleMessageInput(client *Client, serverConnection proto.MessagingServiceClient, input string, stream proto.MessagingService_ChatClient) {
	log.Printf("Lamport timestamp %d, Sending message", client.timestamp)
	err := stream.Send(&proto.ClientSendMessage{ClientName: client.name, Message: input, TimeStamp: int64(client.timestamp)})
	if err != nil {
		log.Fatalln("Could not send message")
	}
}

func handleLeave(client *Client, serverConnection proto.MessagingServiceClient, stream proto.MessagingService_ChatClient) {
	log.Printf("Lamport timestamp %d, Leaving", client.timestamp)
	err := stream.Send(&proto.ClientSendMessage{ClientName: client.name, Message: "leave", TimeStamp: int64(client.timestamp)})

	if err != nil {
		log.Fatalln("Could not leave server")
	}
}

func getServerConnection(c *Client) proto.MessagingServiceClient {
	conn, err := grpc.Dial(":"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("Could not dial server")
	}
	c.updateTimestamp(c.timestamp, &m)
	log.Printf("Lamport timestamp: %d, Joined the server", c.timestamp)
	return proto.NewMessagingServiceClient(conn)
}

func establishConnectionToChat(client *Client, serverConnection proto.MessagingServiceClient) {
	stream, err := serverConnection.Chat(context.Background())
	if err != nil {
		log.Fatalln("could not send join chat")
	}
	waitc := make(chan struct{})
	stream.Send(&proto.ClientSendMessage{ClientName: client.name, Message: "Joined the server", TimeStamp: int64(client.timestamp)})
	go printReceivedMessage(stream, client, waitc)

	readInput(client, serverConnection, stream)
	stream.CloseSend()
	<-waitc
}

func printReceivedMessage(stream proto.MessagingService_ChatClient, c *Client, waitc chan struct{}) {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("client failed: %v", err)
		}
		c.updateTimestamp(int(message.TimeStamp), &m)
		log.Printf("Lamport timestamp: %d, id: %d name: %s, message: %s", c.timestamp, message.Id, message.ClientName, message.Message)
	}
}

func (c *Client) updateTimestamp(newTimestamp int, m *sync.Mutex) {
	m.Lock()
	c.timestamp = syncTimestamp(c.timestamp, newTimestamp)
	c.timestamp++
	m.Unlock()
}

func syncTimestamp(new int, old int) int {
	if new < old {
		return old
	}
	return new
}
