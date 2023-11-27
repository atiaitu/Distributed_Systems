package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin5/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedAuctionServer

	name string // Not required but useful if you want to name your server
	port string // Not required but useful if your server needs to know what port it's listening to

	incrementValue int64      // value that clients can increment.
	mutex          sync.Mutex // used to lock the server to avoid race conditions.

	clients map[gRPC.Auction_BidStreamServer]bool

	timerStarted bool
	auctionTimer *time.Timer
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
var (
	serverName                 = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
	port                       = flag.String("port", "5400", "Server port")     // set with "-port <port>" in terminal
	ServerTimestamp      int64 = 0
	clientsList                = make(map[string]struct{})
	clientConnections    []*grpc.ClientConn
	HighestBid           int64
	CurrentWinner        string
	auctionTimerMutex    sync.Mutex
	AuctionTimerFinished bool
)

func main() {
	// Parse the flags and check if the "wipe" flag is present
	wipeFlag := flag.Bool("wipe", false, "Wipe the log file")
	flag.Parse()

	// Set up the log with the wipeLog condition
	f := setLog(*wipeFlag)
	defer f.Close()

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

func (s *Server) SendBid(ctx context.Context, message *gRPC.BidMessage) (*gRPC.AckAndBid, error) {

	if AuctionTimerFinished {
		return &gRPC.AckAndBid{Message: "The auction is over, use /r to see the winner\n"}, nil

	}
	if int64(message.Bid) <= HighestBid {
		return &gRPC.AckAndBid{Message: "It seems like the current highest bid is above yours. Use /r to check the current highest bid\n"}, nil
	}

	CurrentWinner = message.ClientName
	HighestBid = message.Bid

	// Logs in the terminal when a client sends a message
	log.Printf("%s: Client %s bid: %d with identifier %s", s.name, message.ClientName, message.Bid, message.Identifier) //minus one, because the broadcastchatmessage updates the serverTimestamp after, so we need to correct for that

	// Return an acknowledgment
	return &gRPC.AckAndBid{Message: "bid successfully placed\n"}, nil
}

// BroadcastChatMessage sends an acknowledgment message to all connected clients.
func (s *Server) GetHighestBid(ctx context.Context, NameAndIdentifier *gRPC.NameAndIdentifier) (*gRPC.AckAndBid, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	message := fmt.Sprintf("The current highest bid is: %d\n", HighestBid)

	return &gRPC.AckAndBid{Message: message, Bid: HighestBid}, nil
}

func (s *Server) CheckAuctionState(ctx context.Context, dentifier *gRPC.Identifier) (*gRPC.Ack, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if AuctionTimerFinished {
		message := fmt.Sprintf("Done")

		return &gRPC.Ack{Message: message}, nil
	}

	message := fmt.Sprintf("Alive")

	return &gRPC.Ack{Message: message}, nil
}

func (s *Server) GetWinner(ctx context.Context, dentifier *gRPC.Identifier) (*gRPC.AckAndBid, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	message := fmt.Sprintf("The winner is %s, with the bid: %d\n", CurrentWinner, HighestBid)
	log.Printf("The winner is %s, with the bid: %d\n", CurrentWinner, HighestBid)

	return &gRPC.AckAndBid{Message: message}, nil
}

// Function to add a new client to the list
func (s *Server) addClient(clientName string) {
	s.mutex.Lock()
	clientsList[clientName] = struct{}{}
	s.mutex.Unlock()
}

// Function to handle a new client joining
func (s *Server) HandleNewClient(ctx context.Context, message *gRPC.JoinMessage) (*gRPC.Ack, error) {
	s.addClient(message.Message)

	log.Printf("Participant " + message.Name + " has requested to join the auction")

	// Start the timer when the first client joins
	if !s.timerStarted {
		s.startTimer()
		s.timerStarted = true
	}

	return &gRPC.Ack{Message: "Joined successfully\n"}, nil
}

// function to start the timer
func (s *Server) startTimer() {
	// If there's an existing timer, stop it
	if s.auctionTimer != nil {
		s.auctionTimer.Stop()
	}

	//start a new timer for 30 seconds
	s.auctionTimer = time.NewTimer(30 * time.Second)

	//goroutine to monitor the timer
	go func() {
		<-s.auctionTimer.C
		log.Println("Auction time is up!")

		//use an lock to safely update shared variable
		auctionTimerMutex.Lock()
		AuctionTimerFinished = true
		auctionTimerMutex.Unlock()
	}()
}

// Get preferred outbound ip of this machine
// Useful if you have to know which ip you should dial, in a client running on another computer
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
func setLog(wipeLog bool) *os.File {
	if wipeLog {
		// Clears the log.txt file when wipeLog is true
		if err := os.Truncate("log.txt", 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		}
	}

	// This connects to the log file/changes the output of the log information to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
