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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name       string
	portNumber int
}

var (
	clientPort = flag.Int("clientPort", 8081, "client port number")
	serverPort = flag.Int("serverPort", 8080, "server port number")
	clientName = flag.String("clientName", "DefaultName", "client name")
)

func main() {

	flag.Parse()
	client := &Client{
		name:       *clientName,
		portNumber: *clientPort,
	}
	go startClient(client)
	for {
		time.Sleep(100 * time.Second)
	}
}

func startClient(client *Client) {

	serverConnection := getServerConnection()

	go establishConnectionToChat(client, serverConnection)

	readInput(client, serverConnection)
}

func readInput(client *Client, serverConnection proto.MessagingServiceClient) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		if input == "leave" {
			handleLeave(client, serverConnection)
		} else {
			handleMessageInput(client, serverConnection, input)
		}
	}
}

func handleMessageInput(client *Client, serverConnection proto.MessagingServiceClient, input string) {
	ack, err := serverConnection.SendMessage(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: input})

	if err != nil {
		log.Fatalln("Could not get time")
	}
	log.Printf("Server says %s\n", ack)
}

func handleLeave(client *Client, serverConnection proto.MessagingServiceClient) {
	ack, err := serverConnection.SendMessage(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: "left the server"})
	if err != nil {
		log.Fatalln("Could not leave server")
	}
	log.Printf("Leave: %s\n", ack)
	os.Exit(0)
}

func getServerConnection() proto.MessagingServiceClient {
	conn, err := grpc.Dial(":"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("Could not dial server")
	}
	log.Printf("Joined the server")
	return proto.NewMessagingServiceClient(conn)
}

func establishConnectionToChat(client *Client, serverConnection proto.MessagingServiceClient) {
	stream, err := serverConnection.JoinChat(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: "Joined the server"})
	if err != nil {
		log.Fatalln("could not send join chat")
	}
	printReceivedMessage(stream)
}

func printReceivedMessage(stream proto.MessagingService_JoinChatClient) {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.ListFeatures failed: %v", err)
		}
		log.Printf("Feature: id: %d name: %s, message: %s", message.Id, message.ClientName, message.Message)
	}
}
