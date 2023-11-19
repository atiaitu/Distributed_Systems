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

	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin5/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clientsName             = flag.String("name", "default", "Sender's name")
	serverPort1             = flag.String("server1", "5400", "TCP server 1")
	serverPort2             = flag.String("server2", "5401", "TCP server 2")
	serverPort3             = flag.String("server3", "5402", "TCP server 3")
	server1                 gRPC.AuctionClient
	server2                 gRPC.AuctionClient
	server3                 gRPC.AuctionClient
	ServerConn1             *grpc.ClientConn
	ServerConn2             *grpc.ClientConn
	ServerConn3             *grpc.ClientConn
	stopRoutine             = make(chan bool)
	clientLamportTimestamp  int64
	highestBid              int
	lastGeneratedIdentifier string
	joined                  = false
)

func main() {
	flag.Parse()

	// Connect to the servers and store the clients and connections
	server1, ServerConn1 = ConnectToServer(serverPort1)
	server2, ServerConn2 = ConnectToServer(serverPort2)
	server3, ServerConn3 = ConnectToServer(serverPort3)

	// Close the connections when the program exits
	defer ServerConn1.Close()
	defer ServerConn2.Close()
	defer ServerConn3.Close()

	// Start the binding
	go monitorConnection(ServerConn1, serverPort1)
	go monitorConnection(ServerConn2, serverPort2)
	go monitorConnection(ServerConn3, serverPort3)

	// Start the binding
	parseInput()
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the auctionhouse ")
	fmt.Println("Use /b <your bid> to place a bid")
	fmt.Println("Use /r to review the highest bid or the result")
	fmt.Println("--------------------")

	for {
		fmt.Print("-> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)

		// Generate a unique identifier before sending requests
		identifier := generateUniqueIdentifier()

		// Send requests to all three servers with the same identifier
		response1, err1 := sendRequest(server1, ServerConn1, input, identifier)
		response2, err2 := sendRequest(server2, ServerConn2, input, identifier)
		response3, err3 := sendRequest(server3, ServerConn3, input, identifier)

		// Evaluate responses and take action
		if err1 == nil && err2 == nil && err3 == nil {
			// If all requests are successful, check for a majority response
			majorityResponse := getMajorityResponse(response1, response2, response3)
			handleMajorityResponse(majorityResponse)
		} else {
			// Handle errors if any
			handleError(err1, err2, err3)
		}
	}
}

func sendRequest(server gRPC.AuctionClient, conn *grpc.ClientConn, input, identifier string) (response *gRPC.Ack, err error) {
	if conReady(conn) {
		if !joined {
			sendJoin(server)
			joined = true
		}

		if strings.HasPrefix(input, "/b") {
			// Removing "/b"
			bidAsString := strings.TrimSpace(input[len("/b"):])
			bid, err := strconv.Atoi(bidAsString)
			if err != nil {
				log.Fatal(err)
			}

			var HB, errr = GetHighestBid(server)
			if errr != nil {
				log.Fatal(err)
			}
			if bid < highestBid {
				fmt.Printf("Your bid must be higher than your previous bid\n")
			} else if int64(bid) < HB {
				fmt.Printf("It seems like another bidder has bidden higher. Use /r to check the current highest bid\n")
			} else {
				// Use the same identifier for all multicasted requests
				response, err = sendBid(server, *clientsName, bid, identifier)
				highestBid = bid
			}
		} else if strings.HasPrefix(input, "/r") {
			var response, err = GetHighestBid(server)
			if err != nil {
				log.Printf("%s: Error getting the highest bid: %v", *clientsName, err)
			} else {
				fmt.Printf("The current highest bid is: %d\n", response)
			}
		}
	}

	return response, err
}

func GetHighestBid(server gRPC.AuctionClient) (bid int64, err error) {
	Name := &gRPC.Name{
		Name: *clientsName,
	}
	response, err := server.GetHighestBid(context.Background(), Name)
	if err != nil {
		log.Printf("%s: Error getting the highest bid: %v", *clientsName, err)
		return 0, err
	}

	return response.Bid, nil
}

func sendBid(server gRPC.AuctionClient, clientName string, bid int, identifier string) (response *gRPC.Ack, err error) {
	BidMessage := &gRPC.BidMessage{
		ClientName: clientName,
		Bid:        int64(bid),
		Identifier: identifier,
	}
	return server.SendBid(context.Background(), BidMessage)
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

func sendJoin(server gRPC.AuctionClient) {
	clientLamportTimestamp++

	JoinMessage := &gRPC.JoinMessage{
		Name:       *clientsName,
		Message:    *clientsName + " joined the auction",
		Identifier: generateUniqueIdentifier(),
	}

	// Make a gRPC call to send the join message
	ack, err := server.HandleNewClient(context.Background(), JoinMessage)
	if err != nil {
		log.Printf("Client %s: Error sending join message: %v", *clientsName, err)
		return
	}

	fmt.Printf(ack.Message)
}

func generateUniqueIdentifier() string {
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(1000000)
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	lastGeneratedIdentifier = fmt.Sprintf("%d%d", timestamp, randomNum)
	return lastGeneratedIdentifier
}

func getMajorityResponse(responses ...*gRPC.Ack) *gRPC.Ack {
	// Count the occurrences of each response
	counts := make(map[*gRPC.Ack]int)
	for _, response := range responses {
		if response != nil {
			counts[response]++
		}
	}

	// Find the response with the maximum count
	var majorityResponse *gRPC.Ack
	maxCount := 0
	for response, count := range counts {
		if count > maxCount {
			majorityResponse = response
			maxCount = count
		}
	}

	return majorityResponse
}

func handleMajorityResponse(response *gRPC.Ack) {
	// Handle the majority response here
	if response != nil {
		fmt.Printf("Majority response: %s\n", response.Message)
		// Perform actions based on the majority response
		// ...
	} else {
		fmt.Println("No clear majority response.")
		// Handle the case where there is no clear majority response
		// ...
	}
}

func ConnectToServer(serverPort *string) (gRPC.AuctionClient, *grpc.ClientConn) {
	// Dial options
	// The server is not using TLS, so we use insecure credentials
	// (should be fine for local testing but not in the real world)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Dial the server, with the flag "server", to get a connection to it
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return nil, nil
	}

	// Makes a client from the server connection and saves the connection
	// and prints whether or not the connection was READY
	client := gRPC.NewAuctionClient(conn)

	return client, conn
}

func conReady(conn *grpc.ClientConn) bool {
	if conn != nil {
		return conn.GetState().String() == "READY"
	}
	return false
}

func monitorConnection(conn *grpc.ClientConn, serverPort *string) {
	for {
		// Check if the connection is not ready
		if !conReady(conn) {

			// Attempt to reconnect
			client, newConn := ConnectToServer(serverPort)
			if client != nil {
				// Update the client and connection
				switch serverPort {
				case serverPort1:
					server1 = client
					ServerConn1 = newConn
				case serverPort2:
					server2 = client
					ServerConn2 = newConn
				case serverPort3:
					server3 = client
					ServerConn3 = newConn
				}
			}
		}

		// Sleep for some time before checking again
		time.Sleep(5 * time.Second)
	}
}

func handleError(err1, err2, err3 error) {
	// Handle errors from each server
	// You can customize this based on your requirements
	log.Printf("Error from server 1: %v", err1)
	log.Printf("Error from server 2: %v", err2)
	log.Printf("Error from server 3: %v", err3)
}
