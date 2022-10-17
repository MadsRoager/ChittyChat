package main

import (
	"awesomeProject/proto"
	"bufio"
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"time"
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
		time.Sleep(100)
	}
}

func startClient(client *Client) {

	serverConnection := getServerConnection()
	ack, err := serverConnection.SendMessage(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: "Joined the server"})
	if err != nil {
		log.Fatalln("could not send join message")
	}
	log.Printf("%s", ack)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		log.Printf("Client intputted %s\n", input)

		if input == "leave" {
			ack, err := serverConnection.SendMessage(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: "left the server"})

			if err != nil {
				log.Fatalln("Could not leave server")
			}
			//todo implement leaving
			log.Printf("Leave: %s\n", ack)
			os.Exit(0)
		} else {
			ack, err := serverConnection.SendMessage(context.Background(), &proto.ClientSendMessage{ClientName: client.name, Message: input})

			if err != nil {
				log.Fatalln("Could not get time")
			}

			log.Printf("Server says %s\n", ack)
		}
	}
}

func getServerConnection() proto.MessagingServiceClient {

	conn, err := grpc.Dial(":"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalln("Could not dial server")
	}
	log.Printf("Joined the server")

	return proto.NewMessagingServiceClient(conn)
}
