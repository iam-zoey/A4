package main

import (
	"flag"
	"net"

	"log"
	"net/rpc"
	"sync"
	"time"
)

// Request represents a Seller's request to the Trader
type Request struct {
	SellerID  int
	Post      int
	Item      string
	Quantity  int
	RequestID int // Unique ID for each request
}

// Response represents a Trader's response to the Seller
type Response struct {
	Status    string
	Message   string
	RequestID int
	Processed bool // Indicates if the request was processed
}

// Seller struct represents a seller node
type Seller struct {
	ID          int
	Address     string
	TraderAddr  string
	Post        int
	RequestID   int
	RequestLock sync.Mutex
}

// SendRequest sends incremental requests to the Trader
func (s *Seller) SendRequest() {
	s.RequestLock.Lock()
	s.RequestID++ // Increment request ID for each new request
	reqID := s.RequestID
	s.RequestLock.Unlock()

	req := Request{
		SellerID:  s.ID,
		Post:      s.Post,
		Item:      "apples",
		Quantity:  10,
		RequestID: reqID,
	}

	for {
		client, err := rpc.Dial("tcp", s.TraderAddr)
		if err != nil {
			log.Printf("Seller %d: Failed to connect to Trader at %s. Retrying...", s.ID, s.TraderAddr)
			time.Sleep(5 * time.Second) // Retry after a delay
			continue
		}
		defer client.Close()

		var res Response
		err = client.Call("Trader.ReceiveRequest", &req, &res)
		if err != nil {
			log.Printf("Seller %d: Error sending request: %v. Retrying...", s.ID, err)
			time.Sleep(5 * time.Second) // Retry after a delay
			continue
		}

		if res.Processed && res.RequestID == reqID {
			log.Printf("Seller %d: Request %d processed successfully by Trader", s.ID, reqID)
			break
		} else {
			log.Printf("Seller %d: Trader response indicates request %d not processed. Retrying...", s.ID, reqID)
			time.Sleep(5 * time.Second) // Retry after a delay
		}
	}
}

// UpdateLeader updates the Seller's Trader address after failover
func (s *Seller) UpdateLeader(newLeaderAddr string, reply *string) error {
	log.Printf("Seller %d: Updating Trader to new leader at %s", s.ID, newLeaderAddr)
	s.TraderAddr = newLeaderAddr // Update Trader address
	*reply = "Leader updated successfully"
	return nil
}

// StartRPCServer starts the Seller's RPC server to handle leader updates
func StartRPCServer(s *Seller) {
	err := rpc.Register(s)
	if err != nil {
		log.Fatalf("Error registering Seller service: %v", err)
	}

	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatalf("Error starting RPC server on %s: %v", s.Address, err)
	}
	defer listener.Close()

	log.Printf("Seller %d RPC server started at %s", s.ID, s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func main() {
	id := flag.Int("id", 0, "Seller ID")
	address := flag.String("address", "", "Seller Address")
	traderAddr := flag.String("trader", "", "Trader Address")
	post := flag.Int("post", 0, "Post ID")
	flag.Parse()

	if *id == 0 || *address == "" || *traderAddr == "" || *post == 0 {
		log.Fatal("Usage: seller -id=<id> -address=<address> -trader=<trader> -post=<post>")
	}

	seller := &Seller{
		ID:         *id,
		Address:    *address,
		TraderAddr: *traderAddr,
		Post:       *post,
	}

	// Start the Seller's RPC server in a goroutine
	go StartRPCServer(seller)

	// Periodically send requests
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		seller.SendRequest()
	}
}
