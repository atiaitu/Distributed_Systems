package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin5/proto"

	"google.golang.org/grpc"
)

var ServerTimestamp int64 = 0

type Server struct {
	gRPC.UnimplementedAuctionServer

	name string // Not required but useful if you want to name your server
	port string // Not required but useful if your server needs to know what port it's listening to

	incrementValue int64      // value that clients can increment.
	mutex          sync.Mutex // used to lock the server to avoid race conditions.

	clients map[gRPC.Auction_BidStreamServer]bool
}

// Add connected client streams to the map when they join.
func (s *Server) ChatStream(stream gRPC.Auction_BidStreamServer) error {
	if s.clients == nil {
		s.clients = make(map[gRPC.Auction_BidStreamServer]bool)
	}

	s.mutex.Lock()
	s.clients[stream] = true
	s.mutex.Unlock()

	// Remove the client's stream when it disconnects.
	defer func() {
		s.mutex.Lock()
		delete(s.clients, stream)
		s.mutex.Unlock()
	}()

	// Receive and broadcast messages from the stream.
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		// Process the received message (e.g., log it).
		log.Printf("Received bid from %s: %d", message.ClientName, message.Bid)
	}
	return nil
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal
var clientsList = make(map[string]struct{})
var clientConnections []*grpc.ClientConn
var HighestBid int64

func main() {

	f := setLog() //uncomment this line to log to a log.txt file instead of the console
	defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// launch the server
	launchServer()
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		name:           *serverName,
		port:           *port,
		incrementValue: 0, // gives default value, but not sure if it is necessary
	}

	gRPC.RegisterAuctionServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) SendBid(ctx context.Context, message *gRPC.BidMessage) (*gRPC.Ack, error) {

	HighestBid = message.Bid

	// Logs in the terminal when a client sends a message
	log.Printf("Client %s bid: %d with identifier %s", message.ClientName, message.Bid, message.Identifier) //minus one, because the broadcastchatmessage updates the serverTimestamp after, so we need to correct for that

	// Return an acknowledgment
	return &gRPC.Ack{Message: "bid successfully placed\n"}, nil
}

// BroadcastChatMessage sends an acknowledgment message to all connected clients.
func (s *Server) GetHighestBid(ctx context.Context, name *gRPC.Name) (*gRPC.AckAndBid, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	message := fmt.Sprintf("The current highest bid is: %d\n", HighestBid)

	return &gRPC.AckAndBid{Message: message, Bid: HighestBid}, nil
}

// Function to add a new client to the list
func (s *Server) addClient(clientName string) {
	s.mutex.Lock()
	clientsList[clientName] = struct{}{}
	s.mutex.Unlock()
}

// Function to remove a client from the list
func (s *Server) removeClient(clientName string) {
	s.mutex.Lock()
	delete(clientsList, clientName)
	s.mutex.Unlock()
}

// Function to handle a new client joining
func (s *Server) HandleNewClient(ctx context.Context, message *gRPC.JoinMessage) (*gRPC.Ack, error) {
	s.addClient(message.Message)

	log.Printf("Participant " + message.Name + " have requested to join the auction")

	return &gRPC.Ack{Message: "Joined succesfully\n"}, nil //ikke færdig, skal implementere lamport
}

// Get preferred outbound ip of this machine
// Usefull if you have to know which ip you should dial, in a client running on an other computer
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
