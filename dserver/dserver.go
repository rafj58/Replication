package main

// after 3 strike on ping (every 8 seconds), new master is elected

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	proto "Replication/grpc"
)

type DServer struct {
	proto.UnimplementedAuctionServiceServer
	proto.UnimplementedDistributedServiceServer
	name          string
	address       string
	port          int
	id            int
	mutex         sync.Mutex
	mutex_auction sync.Mutex
}

var (
	my_row   = flag.Int("row", 1, "Indicate the row of parameter file for this peer") // set with "-row <port>" in terminal
	name     = flag.String("name", "peer", "name of the peer")
	confFile = "confFile.csv"
	// store  tcp connection to others peers of the distributed server - id is the key -
	peers = make(map[int]proto.DistributedServiceClient)
	// wait for listen to be open
	wg sync.WaitGroup
	// master info
	masterAddr string
	masterPort int
	masterId   int
	// notifies receipt of a coordination message
	coordination_msg bool
	auction_closed   = false
	auction_amount   = 0
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Missing row parameter, use -row")
		return
	}

	// read from confFile.csv and prepare TCP connections
	csvFile, err := os.Open(confFile)
	if err != nil {
		fmt.Printf("Error while opening CSV file: %v\n", err)
		return
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error in reading CSV file: %v\n", err)
		return
	}

	dserver := &DServer{}
	found := false
	for index, row := range rows {
		// retrive data from csv
		peerAddress := row[0]
		peerPort, _ := strconv.Atoi(row[1])
		peerId, _ := strconv.Atoi(row[2])
		if index == *my_row {
			dserver = &DServer{
				name:    *name,
				address: peerAddress,
				port:    peerPort,
				id:      peerId,
			}
			found = true
		} else {
			// retrieve connection
			connection := connectToPeer(peerAddress, peerPort)
			// add to map id - connection
			peers[peerId] = connection
		}
	}

	if !found {
		fmt.Printf("Row with parameters not founded\n")
		return
	}

	// wait for opening port to listen
	wg.Add(1)

	// open the port to new connections
	go StartListen(dserver)

	wg.Wait()

	// before update the new master i need to update the values of the auction
	// if i will be the new master i will loose this value if i don't do it
	updateActionValues()

	// update the master
	updateMaster(dserver)

	// go routine to check if the master is available
	go checkMaster(dserver)

	for {
		var text string
		log.Print("Insert 'exit' to terminate peer or 'close' to finish the action: ")
		fmt.Scanln(&text)

		if text == "exit" {
			break
		}

		// only the master can close the auction
		if text == "close" {
			if iAmMaster(dserver) {
				auction_closed = true
				updatePeers()
			} else {
				log.Print("Only the master can close the auction")
			}
		}
	}
}

func StartListen(dserver *DServer) {
	// create a new grpc server
	grpcDServer := grpc.NewServer()

	// make the peer listen at the given port
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", dserver.address, strconv.Itoa(dserver.port)))

	if err != nil {
		log.Fatalf("Could not create the peer %v", err)
	}
	log.Printf("Started peer receiving at address: %s and at port: %d\n", dserver.address, dserver.port)

	// register the grpc services
	proto.RegisterDistributedServiceServer(grpcDServer, dserver)
	proto.RegisterAuctionServiceServer(grpcDServer, dserver)

	wg.Done()

	serveError := grpcDServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func connectToPeer(address string, port int) proto.DistributedServiceClient {
	// Dial simply prepare TCP connection
	conn, err := grpc.Dial(address+":"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect to peer %s at port %d", address, port)
	} else {
		log.Printf("Created TCP connection to the %s address at port %d\n", address, port)
	}
	return proto.NewDistributedServiceClient(conn)
}

/*----  Published services for client  ----*/

func (ds *DServer) AskForMaster(ctx context.Context, in *proto.Empty) (*proto.Peer, error) {
	// return the current master
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	return &proto.Peer{
		Address: masterAddr,
		Port:    int32(masterPort),
		Id:      int32(masterId)}, nil
}

func (ds *DServer) Bid(ctx context.Context, in *proto.Amount) (*proto.Empty, error) {
	if !iAmMaster(ds) {
		return nil, status.Errorf(codes.PermissionDenied, "I'am not the master")
	} else {
		ds.mutex_auction.Lock()
		defer ds.mutex_auction.Unlock()
		if !auction_closed {
			if auction_amount > int(in.Amount) { // offer is smaller than current amount
				return nil, status.Errorf(codes.InvalidArgument, "The amount of the bid is too small")
			}
			auction_amount = int(in.Amount)
			log.Printf("Current highest BID is %d", auction_amount)
			updatePeers()
			return &proto.Empty{}, nil
		}
		return nil, status.Errorf(codes.DeadlineExceeded, "The auction is already closed")
	}
}

func (ds *DServer) Result(ctx context.Context, in *proto.Empty) (*proto.Auction, error) {
	if !iAmMaster(ds) {
		return nil, status.Errorf(codes.PermissionDenied, "I'am not the master")
	} else {
		return &proto.Auction{Highest: int32(auction_amount), Closed: auction_closed}, nil
	}
}

/*------ Published services for other peers ----*/

func (ds *DServer) Election(ctx context.Context, in *proto.Peer) (*proto.Empty, error) {
	if ds.id < int(in.Id) {
		return nil, status.Errorf(codes.InvalidArgument, "Peer %d has a lower ID than peer %d", ds.id, in.Id)
	}
	log.Printf("Peer [%s:%d, %d] received election message from [%s:%d, %d]\n", ds.address, ds.port, ds.id, in.Address, in.Port, in.Id)
	// send election to bigger ID than me and reply with ack
	updateMaster(ds)
	return &proto.Empty{}, nil
}

func (ds *DServer) Coordinator(ctx context.Context, in *proto.Peer) (*proto.Empty, error) {
	log.Printf("Peer [%s:%d, %d] is the master.\n", in.Address, in.Port, in.Id)
	coordination_msg = true
	ds.mutex.Lock()
	masterAddr = in.Address
	masterPort = int(in.Port)
	masterId = int(in.Id)
	ds.mutex.Unlock()
	return &proto.Empty{}, nil
}

func (ds *DServer) Ping(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, nil
}

func (ds *DServer) UpdateAuction(ctx context.Context, in *proto.Auction) (*proto.Empty, error) {
	if iAmMaster(ds) {
		log.Print("Something went wrong.. i receive the update auction but i'm the master!")
		return nil, status.Errorf(codes.PermissionDenied, "You can't update the master!")
	}
	log.Printf("Received new value for the action: %d amount, closed %t", in.Highest, in.Closed)
	auction_amount = int(in.Highest)
	auction_closed = in.Closed
	return &proto.Empty{}, nil
}

func (ds *DServer) GetAuctionData(ctx context.Context, in *proto.Empty) (*proto.Auction, error) {
	if iAmMaster(ds) {
		return &proto.Auction{
			Highest: int32(auction_amount),
			Closed:  auction_closed,
		}, nil
	}
	return nil, status.Errorf(codes.Aborted, "Ask to the master for updatest value!")
}

/*--- Function to implement BULLY algorithm ---*/

func updateMaster(ds *DServer) {
	me := &proto.Peer{
		Address: ds.address,
		Port:    int32(ds.port),
		Id:      int32(ds.id),
	}
	coordination_msg = false

	if !sendElectionToBigger(ds, me) {
		// i'am the master
		ds.mutex.Lock()
		masterAddr = ds.address
		masterPort = ds.port
		masterId = ds.id
		ds.mutex.Unlock()
		sendCoordinatorToLower(ds, me)
		log.Printf("I'am the current master!\n")
	} else {
		// wait for coordination message
		time.Sleep(5 * time.Second)
		if !coordination_msg {
			// retry
			log.Printf("No one of bigger id send a coordination message\n")
			updateMaster(ds)
		}
	}
}

// election message to bigger id peers
func sendElectionToBigger(ds *DServer, msg *proto.Peer) bool {
	found_bigger_active := false

	for index, conn := range peers {
		if index > ds.id {
			_, err := conn.Election(context.Background(), msg)
			if err == nil {
				found_bigger_active = true
			}
		}
	}
	return found_bigger_active
}

// coordination to lower if i'm the master
func sendCoordinatorToLower(ds *DServer, msg *proto.Peer) {
	for index, conn := range peers {
		if index < ds.id {
			conn.Coordinator(context.Background(), msg)
		}
	}
}

/* --- functions to monitor master status ---*/

func iAmMaster(ds *DServer) bool {
	return (ds.address == masterAddr && ds.port == masterPort && ds.id == masterId)
}

func checkMaster(ds *DServer) {
	strike := 0
	for {
		select {
		case <-time.NewTicker(8 * time.Second).C:
			if !iAmMaster(ds) {
				if !pingMaster() {
					// master doesn't reply to ping
					strike++
					log.Printf("%d ping lost", strike)
				} else {
					strike = 0
				}
				if strike > 2 {
					// 3 strike in a row, master is no more available
					log.Printf("Master is not available")
					updateMaster(ds)
					strike = 0
				}
			}
		}
	}
}

func pingMaster() bool {
	_, err := peers[masterId].Ping(context.Background(), &proto.Empty{})
	if err != nil {
		return false
	}
	return true
}

// retrive current values of auction
func updateActionValues() {
	for _, conn := range peers {
		msg, err := conn.GetAuctionData(context.Background(), &proto.Empty{})
		if err != nil {
			continue
		}
		log.Printf("Received auction data from master, amount is %d", msg.Highest)
		auction_amount = int(msg.Highest)
		auction_closed = msg.Closed
		break
	}
}

// update others peers in server with auction data
func updatePeers() {
	for index, conn := range peers {
		if index != masterId {
			conn.UpdateAuction(context.Background(), &proto.Auction{
				Highest: int32(auction_amount),
				Closed:  auction_closed})
		}
	}
}
