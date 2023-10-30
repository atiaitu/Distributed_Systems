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
var stopRoutine = make(chan bool)

func main() {
	// Parse flag/arguments
	flag.Parse()

	// Log to file instead of console
	// f := setLog()
	// defer f.Close()

	// Connect to the server and close the connection when the program exits
	ConnectToServer()
	defer ServerConn.Close()

	// Start the binding
	var joined = false

	parseInput(joined)
}

func receiveMessages(chatStream gRPC.Chittychat_ChatStreamClient, stopRoutine chan bool) {
	for {
		select {
		case <-stopRoutine:
			return
		default:
			if chatStream == nil {
				// Ensure that chatStream is not nil before accessing it
				return
			}
			ack, err := chatStream.Recv()
			if ack.ClientName != *clientsName {
				if err != nil {
					log.Printf("Error receiving message from server: %v", err)
					return
				}

				log.Printf("%s: %s", ack.ClientName, ack.Message)
			}
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

	log.Printf(ack.Message)
}

func sendLeaveMessage() {
	client := &gRPC.ClientName{
		ClientName: *clientsName,
	}

	// Make a gRPC call to send the chat message
	ack, err := server.HandleClientLeave(context.Background(), client)
	if err != nil {
		log.Printf("Client %s: Error sending join message: %v", *clientsName, err)
		return
	}
	if ack.Message == "hej" {
	} //so the program can run
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

func parseInput(joined bool) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to Chitty-chat")
	fmt.Println("Use /j to join the chatroom")
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
			if joined {
				// Extract the message text after "/message"
				message := strings.TrimSpace(input[len("/m"):])
				sendChatMessage(*clientsName, message)

			} else {
				log.Println("You cannot write a message before joining. Use /j to join the chat.")
			}
		} else if strings.HasPrefix(input, "/l") {
			if joined {
				// Send leave message
				sendLeaveMessage()
				joined = false
				stopRoutine <- true
			} else {
				log.Println("You cannot leave before joining. Use /j to join the chat.")
			}
		} else if strings.HasPrefix(input, "/j") {
			// Create the bidirectional streaming RPC for message reception.
			chatStream, err := server.ChatStream(context.Background())

			// Start a goroutine to receive and process messages from the server.
			joined = true
			go receiveMessages(chatStream, stopRoutine)
			if err != nil {
				log.Printf("Client %s: Error creating chat stream: %v", *clientsName, err)
				return
			}

			// Send a join message to the server
			joinMessage := fmt.Sprintf("Participant %s joined Chitty-Chat at Lamport time %s", *clientsName, *serverPort) //skal ændres til lamporttime istedet for serverport
			sendJoinMessage(joinMessage)
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
			log.Printf("%s: Error sending chat message: %v", clientName, err)
			return
		}
		log.Printf(ack.Message)
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
