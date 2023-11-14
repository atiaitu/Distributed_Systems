package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atiaitu/Distributed_Systems/tree/main/Handin4/proto"
	gRPC "github.com/atiaitu/Distributed_Systems/tree/main/Handin4/proto"

	"google.golang.org/grpc"
)

type Peer struct {
	ID          string
	Port        string
	server      *grpc.Server
	peerList    []string
	peers       map[string]*Peer
	HasToken    bool
	WantsAccess bool
	mu          sync.Mutex
	proto.PeerToPeerServer
}

var peerName = flag.String("name", "default", "Senders name")
var port = flag.String("port", "5400", "Peer port")

func main() {
	flag.Parse()

	StartPeer()
}

func StartPeer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *peerName, *port)

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *peerName, *port, err)
		return
	}

	var opts []grpc.ServerOption
	grpcPeer := grpc.NewServer(opts...)

	peer := &Peer{
		ID:          *peerName,
		Port:        *port,
		HasToken:    false,
		WantsAccess: false,
		peers:       make(map[string]*Peer),
	}

	gRPC.RegisterPeerToPeerServer(grpcPeer, peer)

	log.Printf("Server %s: Listening at %v\n", *peerName, list.Addr())

	if *port == "5001" {
		peer.HasToken = true
	}

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter text: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if strings.HasPrefix(text, "/add") {
				newPort := text[len("/add"):]
				newPeer := &Peer{
					ID:          fmt.Sprintf("Peer-%s", newPort),
					Port:        newPort,
					HasToken:    false,
					WantsAccess: false,
					peers:       make(map[string]*Peer),
				}
				peer.AddToPeerList(newPeer)
			} else if strings.HasPrefix(text, "/start") {
				StartTokenRing(peer)
			}
		}
	}()

	go func() {
		for {
			// Listen for incoming messages
			// You might want to implement message handling logic here
			time.Sleep(2 * time.Second)
		}
	}()

	if err := grpcPeer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func StartTokenRing(peer *Peer) {
	go func() {
		for {
			peer.CheckIfIWantAccessToCriticalSection()
			time.Sleep(3 * time.Second)
		}
	}()

	go func() {
		for {
			if peer.WantsAccess {
				if peer.HasToken {
					log.Printf("I'm peer %s, and I'm in the critical section :D", *&peer.ID)
					peer.WantsAccess = false
					peer.HasToken = false
					GiveTokenForward(peer)
				}
			} else {
				if peer.HasToken {
					log.Printf("I'm peer %s, and I have the Token", *port)
					peer.HasToken = false
					GiveTokenForward(peer)
				}
			}
		}
	}()
}

func GiveTokenForward(peer *Peer) {
	nextPort := getNextPort(peer.Port)
	if peer.portExists(nextPort) {
		targetPeer := peer.findPeerByPort(nextPort)
		log.Printf(targetPeer.Port)
		if targetPeer != nil {
			peer.SendTokenMessage(nextPort)
			fmt.Printf("Token given to peer %s\n", targetPeer.Port)
		}
	} else if !peer.portExists(nextPort) {
		lowestPortPeer := peer.findLowestPortPeer()
		if lowestPortPeer != nil {
			peer.SendTokenMessage(lowestPortPeer.Port)
			fmt.Printf("Token given to peer %s\n", lowestPortPeer.Port)
		}
	} else {
		GiveTokenForward(peer)
	}
}

func (p *Peer) HandleIncomingMessages(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read the incoming message
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		// Handle the received message (you may want to add more logic here)
		log.Printf("Received message: %s", message)
	}
}

func (p *Peer) SendMessage(ctx context.Context, req *proto.MessageRequest) (*proto.MessageResponse, error) {
	// Handle the received message
	content := req.Content
	log.Printf("Received message: %s", content)

	p.HasToken = true
	time.Sleep(1 * time.Second)
	// Return a response
	return &proto.MessageResponse{Reply: "Message received!"}, nil
}

func (p *Peer) SendTokenMessage(port string) {
	// Create a gRPC client to connect to the target peer
	targetAddr := fmt.Sprintf("localhost:%s", port)
	conn, err := grpc.Dial(targetAddr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Error connecting to peer %s: %v", port, err)
		return
	}
	defer conn.Close()

	// Create a gRPC client
	client := gRPC.NewPeerToPeerClient(conn)

	// Prepare and send the message
	tokenMessageStr := "TOKEN_MESSAGE"
	response, err := client.SendMessage(context.Background(), &gRPC.MessageRequest{Content: tokenMessageStr})
	if err != nil {
		log.Printf("Error sending token message to peer %s: %v", port, err)
		return
	}

	fmt.Printf("Received response from peer %s: %s\n", port, response.Reply)
}

func getNextPort(currentPort string) string {
	portNum := parsePort(currentPort)
	nextPortNum := portNum + 1
	return fmt.Sprintf("%d", nextPortNum)
}

func parsePort(port string) int {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Error parsing port number: %v", err)
		return 0
	}
	return portNum
}

func (p *Peer) portExists(port string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, peerPort := range p.peerList {
		if peerPort == port {
			return true
		}
	}
	return false
}

func (p *Peer) findPeerByPort(port string) *Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, peerPort := range p.peerList {
		if peer, ok := p.peers[peerPort]; ok {
			return peer
		}
	}
	return nil
}

func (p *Peer) findLowestPortPeer() *Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.peerList) == 0 {
		return nil
	}

	lowestPort := p.peerList[0]
	for _, peerPort := range p.peerList {
		if peerPort < lowestPort {
			lowestPort = peerPort
		}
	}

	if lowestPortPeer, ok := p.peers[lowestPort]; ok {
		return lowestPortPeer
	}
	return nil
}

func (p *Peer) StartCheckingAccess() {
	go func() {
		for {
			p.CheckIfIWantAccessToCriticalSection()
			time.Sleep(10 * time.Second)
		}
	}()
}

func (p *Peer) CheckIfIWantAccessToCriticalSection() {
	if rollDice() == 6 {
		p.IWantAccess()
	}
}

func rollDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}

func (p *Peer) IWantAccess() {
	if !p.WantsAccess {
		fmt.Printf("Peer %s wants access to the critical section!\n", p.ID)
		p.WantsAccess = true
	}
}

func (p *Peer) AddToPeerList(newPeer *Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peerList = append(p.peerList, newPeer.Port)
	p.peers[newPeer.Port] = newPeer

	fmt.Printf("Node %s added to peer list: %v\n", newPeer.ID, p.peerList)
}

func (p *Peer) Stop() {
	p.server.Stop()
}
