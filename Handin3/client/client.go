package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChittychatClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {
	// Parse flag/arguments
	flag.Parse()

	fmt.Println("--- Chitty-Chat ---")

	// Log to file instead of console
	// f := setLog()
	// defer f.Close()

	// Connect to the server and close the connection when the program exits
	ConnectToServer()
	defer ServerConn.Close()

	// Create the bidirectional streaming RPC for message reception.
	chatStream, err := server.ChatStream(context.Background())
	if err != nil {
		log.Printf("Client %s: Error creating chat stream: %v", *clientsName, err)
		return
	} else {
		log.Printf("Client %s joined on chat stream", *clientsName)
	}

	// Start a goroutine to receive and process messages from the server.
	go receiveMessages(chatStream)

	// Send a join message to the server
	joinMessage := fmt.Sprintf("%s joined the server on port %s", *clientsName, *serverPort)
	sendJoinMessage(joinMessage)

	// Start the binding
	parseInput()

}

func receiveMessages(chatStream gRPC.Chittychat_ChatStreamClient) {
	for {
		ack, err := chatStream.Recv()
		if ack.ClientName != *clientsName {
			if err != nil {
				log.Printf("Error receiving message from server: %v", err)
				return
			}

			log.Printf("Message from %s: %s", ack.ClientName, ack.Message)
		}
	}
}

func sendJoinMessage(message string) {
	chatMessage := &gRPC.JoinMessage{
		Name:    *clientsName,
		Message: message,
	}

	// Make a gRPC call to send the chat message
	ack, err := server.HandleNewClient(context.Background(), chatMessage)
	if err != nil {
		log.Printf("Client %s: Error sending join message: %v", *clientsName, err)
		return
	}

	log.Printf("Client %s: %s", *clientsName, ack.Message)
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewChittychatClient(conn)
	ServerConn = conn
	log.Println("the connection is:", conn.GetState().String())
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to Chitty-chat")
	fmt.Println("Use /m <your message> to write a message")
	fmt.Println("Use /l to leave the chatroom")
	fmt.Println("--------------------")

	for {
		fmt.Print("-> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		if strings.HasPrefix(input, "/m") {
			// Extract the message text after "/message"
			message := strings.TrimSpace(input[len("/m"):])
			sendChatMessage(*clientsName, message)

		} else {
			val, err := strconv.ParseInt(input, 10, 64)
			if err != nil {
				log.Printf("Invalid input: %s", input)
				continue
			}
			incrementVal(val)
		}
	}
}

func incrementVal(val int64) {
	//create amount type
	amount := &gRPC.Amount{
		ClientName: *clientsName,
		Value:      val, //cast from int to int32
	}

	//Make gRPC call to server with amount, and recieve acknowlegdement back.
	ack, err := server.Increment(context.Background(), amount)
	if err != nil {
		log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
		log.Println(err)
	}

	// check if the server has handled the request correctly
	if ack.NewValue >= val {
		fmt.Printf("Success, the new value is now %d\n", ack.NewValue)
	} else {
		// something could be added here to handle the error
		// but hopefully this will never be reached
		fmt.Println("Oh no something went wrong :(")
	}
}

func sendChatMessage(clientName string, message string) {
	chatMessage := &gRPC.ChatMessage{
		ClientName: clientName,
		Message:    message,
	}

	// Make a gRPC call to send the chat message
	if len(chatMessage.Message) > 128 {
		log.Printf("Client %s: The length of your message must be under 128 characters", clientName)
	} else {
		ack, err := server.SendChatMessage(context.Background(), chatMessage)
		if err != nil {
			log.Printf("Client %s: Error sending chat message: %v", clientName, err)
			return
		}
		log.Printf("Client %s: %s", clientName, ack.Message)
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.ChittychatClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
