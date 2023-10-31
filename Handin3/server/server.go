package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin3/proto"

	"google.golang.org/grpc"
)

var ServerTimestamp int64 = 0

type Server struct {
	gRPC.UnimplementedChittychatServer

	name string // Not required but useful if you want to name your server
	port string // Not required but useful if your server needs to know what port it's listening to

	incrementValue int64      // value that clients can increment.
	mutex          sync.Mutex // used to lock the server to avoid race conditions.

	clients map[gRPC.Chittychat_ChatStreamServer]bool
}

// Add connected client streams to the map when they join.
func (s *Server) ChatStream(stream gRPC.Chittychat_ChatStreamServer) error {
	if s.clients == nil {
		s.clients = make(map[gRPC.Chittychat_ChatStreamServer]bool)
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

	gRPC.RegisterChittychatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) SendChatMessage(ctx context.Context, message *gRPC.ChatMessage) (*gRPC.Ack, error) {

	// Broadcast the message to all connected clients
	s.BroadcastChatMessage(message)

	// Logs in the terminal when a client sends a message
	log.Printf("Client %s sent message: %s at Lamport time L %v", message.ClientName, message.Message, ServerTimestamp-1) //minus one, because the broadcastchatmessage updates the serverTimestamp after, so we need to correct for that

	// Return an acknowledgment
	return &gRPC.Ack{Message: "Chat message sent successfully ", Timestamp: ServerTimestamp - 1}, nil
}

// BroadcastChatMessage sends an acknowledgment message to all connected clients.
func (s *Server) BroadcastChatMessage(message *gRPC.ChatMessage) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var time1 int64 = max(message.Timestamp, ServerTimestamp)
	var servertimeStamp string = strconv.FormatInt(time1, 10)

	log.Printf("Participant " + message.ClientName + " have requested to requested to publish a message at Lamport time L " + servertimeStamp)

	var time2 int64 = max(message.Timestamp, ServerTimestamp) + 1
	var serverTimeStamp string = strconv.FormatInt(time2, 10)
	ServerTimestamp = time2 + 1

	JoinMessage := &gRPC.ChatMessage{
		Message: message.ClientName + " published a message at Lamport time L " + serverTimeStamp + " : " + message.Message,
	}

	for clientStream := range s.clients {
		if err := clientStream.Send(JoinMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}
}

// Function to add a new client to the list
func (s *Server) addClient(clientName string) {
	s.mutex.Lock()
	clientsList[clientName] = struct{}{}
	s.mutex.Unlock()
}

func (s *Server) BroadcastJoinedClient(clientName string, timestamp int64) gRPC.Ack {

	var time1 int64 = max(timestamp, ServerTimestamp)
	var servertimeStamp string = strconv.FormatInt(time1, 10)

	log.Printf("Participant " + clientName + " have requested to join at Lamport time L " + servertimeStamp)

	var time2 int64 = max(timestamp, ServerTimestamp) + 1
	var serverTimeStamp string = strconv.FormatInt(time2, 10)
	ServerTimestamp = time2 + 1

	JoinMessage := &gRPC.ChatMessage{
		Message: "Participant " + clientName + " joined Chitty-Chat at Lamport time L " + serverTimeStamp,
	}

	for clientStream := range s.clients {
		if err := clientStream.Send(JoinMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}

	log.Printf(JoinMessage.Message)

	return gRPC.Ack{Message: JoinMessage.Message, Timestamp: timestamp}
}

func (s *Server) BroadcastClientLeave(clientName string, timestamp int64) gRPC.Ack {

	var time1 int64 = max(timestamp, ServerTimestamp)
	var servertimeStamp string = strconv.FormatInt(time1, 10)

	log.Printf("Participant " + clientName + " have requested to requested to leave at Lamport time L " + servertimeStamp)

	var time2 int64 = max(timestamp, ServerTimestamp) + 1
	var serverTimeStamp string = strconv.FormatInt(time2, 10)
	ServerTimestamp = time2 + 1

	LeaveMessage := &gRPC.ChatMessage{
		Message: "Participant " + clientName + " left Chitty-Chat at Lamport time L " + serverTimeStamp,
	}

	for clientStream := range s.clients {
		if err := clientStream.Send(LeaveMessage); err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Handle the error, e.g., remove the disconnected client from the list.
		}
	}

	log.Printf(LeaveMessage.Message)

	return gRPC.Ack{Message: LeaveMessage.Message, Timestamp: timestamp}
}

// Function to remove a client from the list
func (s *Server) removeClient(clientName string) {
	s.mutex.Lock()
	delete(clientsList, clientName)
	s.mutex.Unlock()
}

// Function to handle a new client joining
func (s *Server) HandleNewClient(ctx context.Context, message *gRPC.JoinOrLeaveMessage) (*gRPC.GiveTimestampAndAck, error) {
	s.addClient(message.Message)
	var clientJoinMessage = s.BroadcastJoinedClient(message.Name, message.Timestamp)

	return &gRPC.GiveTimestampAndAck{Message: clientJoinMessage.Message, Timestamp: clientJoinMessage.Timestamp}, nil //ikke færdig, skal implementere lamport
}

// Function to handle a client leaving
func (s *Server) HandleClientLeave(ctx context.Context, LeaveMessage *gRPC.JoinOrLeaveMessage) (*gRPC.GiveTimestampAndAck, error) {
	s.removeClient(LeaveMessage.Message)
	var clientLeaveMessage = s.BroadcastClientLeave(LeaveMessage.Name, LeaveMessage.Timestamp)

	return &gRPC.GiveTimestampAndAck{Message: clientLeaveMessage.Message}, nil
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
