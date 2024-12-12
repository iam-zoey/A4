package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// ======= STRUCTS =======
type Trader struct {
	ID          int
	Address     string
	Peer        string
	Post        int
	IsLeader    bool
	Heartbeat   bool
	HeartbeatMu sync.Mutex
	Requests    []Request
	RequestMu   sync.Mutex
}

type Response struct {
	Status    string
	Message   string
	RequestID int
	Processed bool // Indicates if the request was processed
}

type Request struct {
	SellerID  int
	Post      int
	Item      string
	Quantity  int
	RequestID int // Unique ID for each request
}

// ForwardRequest forwards the request to the peer Trader
func (t *Trader) ForwardRequest(req *Request) {
	client, err := rpc.Dial("tcp", t.Peer)
	if err != nil {
		log.Printf("Trader %d: Failed to connect to peer Trader at %s to forward request.", t.ID, t.Peer)
		return
	}
	defer client.Close()

	var reply string
	err = client.Call("Trader.ReceiveRequest", req, &reply)
	if err != nil {
		log.Printf("Trader %d: Failed to forward request: %v", t.ID, err)
		return
	}

	log.Printf("Trader %d: Request forwarded successfully to Trader %s", t.ID, t.Peer)
}

// ReceiveHeartbeat handles heartbeat messages from the peer Trader
func (t *Trader) ReceiveHeartbeat(req int, reply *string) error {
	t.HeartbeatMu.Lock()
	t.Heartbeat = true
	t.HeartbeatMu.Unlock()

	log.Printf("Trader %d: Received heartbeat from Trader %d", t.ID, req)
	*reply = "Alive"
	return nil
}

// SendHeartbeat sends heartbeat messages to the peer Trader
func (t *Trader) SendHeartbeat() {
	client, err := rpc.Dial("tcp", t.Peer)
	if err != nil {
		log.Printf("Trader %d: Failed to connect to peer Trader at %s. Assuming failure.", t.ID, t.Peer)
		t.TakeOverLeadership()
		return
	}
	defer client.Close()

	var reply string
	err = client.Call("Trader.ReceiveHeartbeat", t.ID, &reply)
	if err != nil {
		log.Printf("Trader %d: Failed to send heartbeat: %v", t.ID, err)
		t.TakeOverLeadership()
		return
	}

	log.Printf("Trader %d: Heartbeat acknowledged by peer %s", t.ID, t.Peer)
}

// StartHeartbeat sends periodic heartbeat messages to the peer Trader
func (t *Trader) StartHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.SendHeartbeat()
	}
}

// StartRPCServer starts the Trader's RPC server
func StartRPCServer(t *Trader) {
	err := rpc.Register(t)
	if err != nil {
		log.Fatalf("Error registering Trader service: %v", err)
	}

	listener, err := net.Listen("tcp", t.Address)
	if err != nil {
		log.Fatalf("Error starting RPC server on %s: %v", t.Address, err)
	}
	defer listener.Close()

	log.Printf("Trader %d RPC server started at %s", t.ID, t.Address)

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
	id := flag.Int("id", 0, "Trader ID")
	address := flag.String("address", "", "Trader Address")
	peer := flag.String("peer", "", "Peer Trader Address")
	post := flag.Int("post", 0, "Post ID")
	flag.Parse()

	if *id == 0 || *address == "" || *peer == "" || *post == 0 {
		log.Fatal("Usage: trader -id=<id> -address=<address> -peer=<peer> -post=<post>")
	}

	trader := &Trader{
		ID:       *id,
		Address:  *address,
		Peer:     *peer,
		Post:     *post,
		IsLeader: *id == 1, // Assume Trader 1 starts as the leader
	}

	go StartRPCServer(trader)
	go trader.StartHeartbeat()

	select {} // Keep the process running
}

// SendResponse sends a response back to the Seller
func (t *Trader) SendResponse(sellerAddr string, res *Response) {
	client, err := rpc.Dial("tcp", sellerAddr)
	if err != nil {
		log.Printf("Trader %d: Failed to connect to Seller at %s: %v", t.ID, sellerAddr, err)
		return
	}
	defer client.Close()

	var reply string
	err = client.Call("Seller.ReceiveResponse", res, &reply)
	if err != nil {
		log.Printf("Trader %d: Failed to send response to Seller at %s: %v", t.ID, sellerAddr, err)
		return
	}

	log.Printf("Trader %d: Response sent to Seller at %s", t.ID, sellerAddr)
}

// NotifySellers informs all Sellers to communicate with the new leader
func (t *Trader) NotifySellers(newLeaderAddr string) {
	// Simulate a list of seller addresses (in a real system, this would be dynamically populated)
	sellerAddresses := []string{"localhost:8003", "localhost:8004"} // Add all known Seller addresses here

	for _, sellerAddr := range sellerAddresses {
		client, err := rpc.Dial("tcp", sellerAddr)
		if err != nil {
			log.Printf("Trader %d: Failed to notify Seller at %s: %v", t.ID, sellerAddr, err)
			continue
		}
		defer client.Close()

		var reply string
		err = client.Call("Seller.UpdateLeader", newLeaderAddr, &reply)
		if err != nil {
			log.Printf("Trader %d: Failed to notify Seller at %s: %v", t.ID, sellerAddr, err)
			continue
		}

		log.Printf("Trader %d: Notified Seller at %s about new leader %s", t.ID, sellerAddr, newLeaderAddr)
	}
}

// TakeOverLeadership promotes the Trader as the leader for all posts and informs Sellers
func (t *Trader) TakeOverLeadership() {
	t.IsLeader = true
	log.Printf("Trader %d: Taking over all posts as the sole leader.", t.ID)

	// Notify Sellers about the new leader
	t.NotifySellers(t.Address)
}

// ReceiveRequest handles requests from Sellers
func (t *Trader) ReceiveRequest(req *Request, res *Response) error {
	log.Printf("Trader %d: Received request %d from Seller %d for %d %s in Post %d",
		t.ID, req.RequestID, req.SellerID, req.Quantity, req.Item, req.Post)

	// Simulate request processing
	time.Sleep(2 * time.Second)

	res.RequestID = req.RequestID
	res.Status = "Success"
	res.Message = fmt.Sprintf("Processed request %d: %d %s from Seller %d", req.RequestID, req.Quantity, req.Item, req.SellerID)
	res.Processed = true
	return nil
}
