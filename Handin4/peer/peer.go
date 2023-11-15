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
	peerPorts   []string
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

	//open log file
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	// Check if the log file exists
	fileInfo, err := logFile.Stat()
	if err == nil && fileInfo.Size() > 0 {
		// If the log file exists and has content, truncate it
		if err := os.Truncate("log.txt", 0); err != nil {
			log.Fatalf("Error truncating log file: %v", err)
		}
		log.Printf("Log file truncated\n")
	}

	log.SetOutput(logFile)

	fmt.Printf("Peer %s: Attempts to create listener on port %s\n", *peerName, *port)

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

	fmt.Printf("Peer %s: Listening at %v\n", *peerName, list.Addr())

	fmt.Printf("HELLO PEER\n")
	fmt.Printf("/add <port number> to add another peer (service discovery)\n")
	fmt.Printf("/start to start access to tokenring\n")

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter input: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if strings.HasPrefix(text, "/add") {
				newPort := text[len("/add "):]
				newPeer := &Peer{
					ID:          fmt.Sprintf(newPort),
					Port:        newPort,
					HasToken:    false,
					WantsAccess: false,
					peers:       make(map[string]*Peer),
				}
				peer.AddToPeerList(newPeer)
			} else if strings.HasPrefix(text, "/start") {
				//give token to the "first" peer (on port 5001)
				if *port == "5001" {
					peer.HasToken = true
				}
				log.Printf("Peer %s joined the tokenring", peer.ID)
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
					log.Printf("I'm peer %s, and I have the Token", *&peer.ID)
					log.Printf("I'm peer %s, and I'm in the critical section :D", *&peer.ID)
					time.Sleep(2 * time.Second)
					log.Printf("I'm peer %s, and i left the critical section :D", *&peer.ID)
					peer.WantsAccess = false
					peer.HasToken = false
					GiveTokenForward(peer)
				}
			} else {
				if peer.HasToken {
					log.Printf("I'm peer %s, and I have the Token", *&peer.ID)
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
		if targetPeer != nil {
			peer.SendTokenMessage(nextPort)
			fmt.Printf("Token given to peer %s\n", targetPeer.ID)
		}
	} else if !peer.portExists(nextPort) {
		nextPort2 := getNextPort(nextPort)
		if peer.portExists(nextPort2) {
			log.Printf("peer on port next port must be dead, sending to +2 port")
			targetPeer := peer.findPeerByPort(nextPort)
			if targetPeer != nil {
				peer.SendTokenMessage(nextPort)
				fmt.Printf("Token given to peer %s\n", targetPeer.ID)
			}
		} else if !peer.portExists(nextPort2) {
			lowestPortPeer := peer.findLowestPortPeer()
			if lowestPortPeer != nil {
				peer.SendTokenMessage(lowestPortPeer.Port)
				fmt.Printf("Token given to peer %s\n", lowestPortPeer.ID)
			}
		}
	} else {
		GiveTokenForward(peer)
	}
}

func (p *Peer) SendMessage(ctx context.Context, req *proto.MessageRequest) (*proto.MessageResponse, error) {
	// Handle the received message
	content := req.Content
	log.Printf("Peer %s received the %s", p.ID, content)

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
	tokenMessageStr := "TOKEN"
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

	for _, peerPort := range p.peerPorts {
		if peerPort == port {
			return true
		}
	}
	return false
}

func (p *Peer) findPeerByPort(port string) *Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, peerPort := range p.peerPorts {
		if peer, ok := p.peers[peerPort]; ok {
			return peer
		}
	}
	return nil
}

func (p *Peer) findLowestPortPeer() *Peer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.peerPorts) == 0 {
		return nil
	}

	lowestPort := p.peerPorts[0]
	for _, peerPort := range p.peerPorts {
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
		log.Printf("Peer %s wants access to the critical section!\n", p.ID)
		p.WantsAccess = true
	}
}

func (p *Peer) AddToPeerList(newPeer *Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peerPorts = append(p.peerPorts, newPeer.Port)
	p.peers[newPeer.Port] = newPeer

	log.Printf("Port %s added to peer %s's list: %v\n", newPeer.Port, p.ID, p.peerPorts)
}

func (p *Peer) Stop() {
	p.server.Stop()
}
