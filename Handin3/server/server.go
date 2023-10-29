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
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin3/proto"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChittychatServer

	name string // Not required but useful if you want to name your server
	port string // Not required but useful if your server needs to know what port it's listening to

	incrementValue int64      // value that clients can increment.
	mutex          sync.Mutex // used to lock the server to avoid race conditions.

	clients    map[gRPC.Chittychat_ChatStreamServer]bool
	clientsMux sync.Mutex
}

type Client struct {
	conn gRPC.ChittychatClient // Replace with the actual gRPC client type
}

// Initialize the map in your server's constructor.
func NewServer() *Server {
	return &Server{
		clients: make(map[gRPC.Chittychat_ChatStreamServer]bool),
	}
}

// Add connected client streams to the map when they join.
func (s *Server) ChatStream(stream gRPC.Chittychat_ChatStreamServer) error {
	if s.clients == nil {
		s.clients = make(map[gRPC.Chittychat_ChatStreamServer]bool)
	}

	s.clientsMux.Lock()
	s.clients[stream] = true
	s.clientsMux.Unlock()

	// Remove the client's stream when it disconnects.
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients, stream)
		s.clientsMux.Unlock()
	}()

	// Receive and broadcast messages from the stream.
	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		// Process the received message (e.g., log it).
		log.Printf("Received message from %s: %s", message.ClientName, message.Message)
	}
	return nil
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal
var clientsList = make(map[string]struct{})
var clientConnections []*grpc.ClientConn

func main() {

	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// launch the server
	launchServer()

	// code here is unreachable because launchServer occupies the current thread.
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

	gRPC.RegisterChittychatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

// The method format can be found in the pb.go file. If the format is wrong, the server type will give an error.
func (s *Server) Increment(ctx context.Context, Amount *gRPC.Amount) (*gRPC.Ack, error) {
	// locks the server ensuring no one else can increment the value at the same time.
	// and unlocks the server when the method is done.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// increments the value by the amount given in the request,
	// and returns the new value.
	s.incrementValue += int64(Amount.GetValue())
	return &gRPC.Ack{NewValue: s.incrementValue}, nil
}

func (s *Server) SendChatMessage(ctx context.Context, message *gRPC.ChatMessage) (*gRPC.Ack1, error) {
	// Logs in the terminal when a client sends a message
	log.Printf("Client %s sent message: %s", message.ClientName, message.Message)

	// Broadcast the message to all connected clients
	s.BroadcastChatMessage(message)

	// Return an acknowledgment
	return &gRPC.Ack1{Message: "Chat message sent successfully"}, nil
}

// BroadcastChatMessage sends an acknowledgment message to all connected clients.
func (s *Server) BroadcastChatMessage(message *gRPC.ChatMessage) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	for clientStream := range s.clients {
		ackMessage := &gRPC.ChatMessage{Message: message.Message, ClientName: message.ClientName}
		if err := clientStream.Send(ackMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}
}

func (c *Client) SendMessage(message *gRPC.ChatMessage) error {
	_, err := c.conn.SendChatMessage(context.Background(), message)
	if err != nil {
		return err
	}
	return nil
}

// Function to add a new client to the list
func (s *Server) addClient(clientName string) {
	s.mutex.Lock()
	clientsList[clientName] = struct{}{}

	for clientStream := range s.clients {
		ackMessage := &gRPC.ChatMessage{ClientName: clientName}
		if err := clientStream.Send(ackMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}
	s.mutex.Unlock()
}

// Function to remove a client from the list
func (s *Server) removeClient(clientName string) {
	s.mutex.Lock()
	delete(clientsList, clientName)

	for clientStream := range s.clients {
		ackMessage := &gRPC.ChatMessage{ClientName: clientName}
		if err := clientStream.Send(ackMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}

	s.mutex.Unlock()
}

// Function to handle a new client joining
func (s *Server) HandleNewClient(ctx context.Context, message *gRPC.JoinMessage) (*gRPC.Ack1, error) {
	s.addClient(message.Message)
	log.Printf("%s", message.Message)
	// You can perform additional actions here when a new client joins.
	return &gRPC.Ack1{Message: "Join message sent successfully"}, nil
}

// Function to handle a client leaving
func (s *Server) handleClientLeave(clientName string) {
	s.removeClient(clientName)
	// You can perform additional actions here when a client leaves.
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
