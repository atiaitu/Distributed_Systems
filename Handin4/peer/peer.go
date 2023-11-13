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
		for {
			peer.CheckIfIWantAccessToCriticalSection()
			time.Sleep(10 * time.Second)
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

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter text: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if strings.HasPrefix(text, "/add") {
				peer.AddToPeerList(text[len("/add"):])
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

func GiveTokenForward(peer *Peer) {
	nextPort := getNextPort(peer.Port)
	if peer.portExists(nextPort) {
		targetPeer := peer.findPeerByPort(nextPort)
		if targetPeer != nil {
			targetPeer.HasToken = true
			peer.SendTokenMessage(targetPeer)
			fmt.Printf("Token given to peer %s\n", targetPeer.Port)
		}
	} else if !peer.portExists(nextPort) {
		lowestPortPeer := peer.findLowestPortPeer()
		if lowestPortPeer != nil {
			log.Printf("2")
			lowestPortPeer.HasToken = true
			peer.SendTokenMessage(lowestPortPeer)
			fmt.Printf("Token given to peer %s\n", lowestPortPeer.Port)
		}
	} else {
		GiveTokenForward(peer)
	}
}

func (p *Peer) HandleMessage(ctx context.Context, req *gRPC.MessageRequest) (*gRPC.MessageResponse, error) {
	content := req.GetContent()

	// Check if the received message is a token and perform actions accordingly
	if content == "TOKEN_MESSAGE" {
		// Add your logic here to handle the received token message
		// For example, set HasToken to true, perform some critical section, etc.
		p.HasToken = true
		fmt.Printf("Received Token Message from peer %s\n", req.SenderPort)
	}

	reply := fmt.Sprintf("Node %s received: %s", p.ID, content)
	return &gRPC.MessageResponse{Reply: reply}, nil
}

func (p *Peer) SendTokenMessage(targetPeer *Peer) {
	tokenMessageStr := "TOKEN_MESSAGE"
	response, err := targetPeer.HandleMessage(context.Background(), &gRPC.MessageRequest{
		Content:      tokenMessageStr,
		SenderPort:   p.Port,
		ReceiverPort: targetPeer.Port,
	})
	if err != nil {
		fmt.Printf("Error sending token message to peer %s: %v\n", targetPeer.Port, err)
		return
	}

	fmt.Printf("Received response from peer %s: %s\n", targetPeer.Port, response.Reply)
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

func (p *Peer) AddToPeerList(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peerList = append(p.peerList, text)
	fmt.Printf("Node %s added to peer list: %v\n", p.ID, p.peerList)
}

func (p *Peer) Stop() {
	p.server.Stop()
}
