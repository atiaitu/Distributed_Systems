package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin5/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.AuctionClient   //the server
var ServerConn *grpc.ClientConn //the server connection
var stopRoutine = make(chan bool)
var clientlamportTimestamp int64
var highestbid int

func main() {
	// Parse flag/arguments
	flag.Parse()

	//Log to file instead of console
	//f := setLog()
	//defer f.Close()

	// Connect to the server and close the connection when the program exits
	ConnectToServer()
	defer ServerConn.Close()

	// Start the binding
	var joined = false

	parseInput(joined)
}

func parseInput(joined bool) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the auctionhouse ")
	fmt.Println("Use /b <yuor bid> to place a bid")
	fmt.Println("Use /r to review the highest bid or the result")
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

		if strings.HasPrefix(input, "/b") {
			if !joined {
				sendJoin()
				joined = true

			}
			// removing "/m"
			bidAsString := strings.TrimSpace(input[len("/b"):])
			bid, err := strconv.Atoi(bidAsString)
			if err != nil {
				log.Fatal(err)
			}
			if bid < highestbid {
				fmt.Printf("Your bid must be higher than your previous bid\n")
			} else if int64(bid) < GetHighestBid() {
				fmt.Printf("it seems like another bidder has bidden higher, use /r to the the current highest bid\n")
			} else {
				sendBid(*clientsName, bid)
				highestbid = bid
			}
		} else if strings.HasPrefix(input, "/r") {
			PrintHighestBid()
		}
	}
}

func GetHighestBid() int64 {
	Name := &gRPC.Name{
		Name: *clientsName,
	}
	ack, err := server.GetHighestBid(context.Background(), Name)
	if err != nil {
		log.Printf("%s: Error sending chat message: %v", Name.Name, err)
		return int64(highestbid)
	}

	return ack.Bid
}

func PrintHighestBid(){
	Name := &gRPC.Name{
		Name: *clientsName,
	}
	ack, err := server.GetHighestBid(context.Background(), Name)
	if err != nil {
		log.Printf("%s: Error sending chat message: %v", Name.Name, err)
		return
	}

	fmt.Printf(ack.Message)
}

func receiveMessages(chatStream gRPC.Auction_BidStreamClient) {
	for {
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

			log.Printf("Client %s won with the bid: %d", ack.ClientName, ack.Bid)
		}
	}
}

func sendJoin() {

	//increment timestamp before event
	clientlamportTimestamp++

	JoinMessage := &gRPC.JoinMessage{
		Name:       *clientsName,
		Message:    *clientsName + " joined the auction",
		Identifier: generateUniqueIdentifier(),
	}

	// Make a gRPC call to send the chat message
	ack, err := server.HandleNewClient(context.Background(), JoinMessage)
	if err != nil {
		log.Printf("Client %s: Error sending join message: %v", *clientsName, err)
		return
	}

	fmt.Printf(ack.Message)
}

func sendBid(clientName string, bid int) {

	BidMessage := &gRPC.BidMessage{
		ClientName: clientName,
		Bid:        int64(bid),
		Identifier: generateUniqueIdentifier(),
	}

	// Make a gRPC call to send the chat message
	ack, err := server.SendBid(context.Background(), BidMessage)
	if err != nil {
		log.Printf("%s: Error sending chat message: %v", clientName, err)
		return
	}

	fmt.Printf(ack.Message)
}

// from ChatGpt
func generateUniqueIdentifier() string {
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(1000000)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%d%d", timestamp, randomNum)
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
	server = gRPC.NewAuctionClient(conn)
	ServerConn = conn
	log.Println("the connection is:", conn.GetState().String())
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.AuctionClient) bool {
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
