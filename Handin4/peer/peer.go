package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin4/proto"

	"google.golang.org/grpc"
)

// this has to be the same as the go.mod module,
// followed by the path to the folder the proto file is ip.

type Peer struct {
	ID     string
	Port   string
	server *grpc.Server
}

func (p *Peer) SendMessage(ctx context.Context, req *gRPC.MessageRequest) (*gRPC.MessageResponse, error) {
	content := req.GetContent()
	reply := fmt.Sprintf("Node %s received: %s", p.ID, content)
	return &gRPC.MessageResponse{Reply: reply}, nil
}

func (p *Peer) Start() {
	listener, err := net.Listen("tcp", "localhost:"+p.Port)
	if err != nil {
		log.Fatalf("Error creating listener for Node %s: %v", p.ID, err)
	}

	p.server = grpc.NewServer()
	gRPC.RegisterP2PServiceServer(p.server, p)

	fmt.Printf("Node %s listening on %s\n", p.ID, listener.Addr())

	if err := p.server.Serve(listener); err != nil {
		log.Fatalf("Error serving Node %s: %v", p.ID, err)
	}
}

func (p *Peer) Stop() {
	p.server.Stop()
}

func main() {
	node1 := &Peer{ID: "1", Port: "50051"}
	node2 := &Peer{ID: "2", Port: "50052"}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		node1.Start()
	}()

	go func() {
		defer wg.Done()
		node2.Start()
	}()

	// Simulate sending a message from node1 to node2
	node1Client, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Node 2: %v", err)
	}
	defer node1Client.Close()

	client := gRPC.NewP2PServiceClient(node1Client)
	message := "Hello, Node 2!"

	response, err := client.SendMessage(context.Background(), &gRPC.MessageRequest{Content: message})
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	fmt.Printf("Response from Node 2 to Node 1: %s\n", response.GetReply())

	// Stop the nodes
	node1.Stop()
	node2.Stop()

	wg.Wait()
}
