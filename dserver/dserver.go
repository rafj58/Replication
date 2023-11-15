package main

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
	name    string
	address string
	port    int
	id      int
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
	coordination_msg chan bool
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

	findMaster(dserver)

	go checkMaster(dserver)

	for {
		var text string
		log.Printf("Insert 'exit' to quit\n")
		fmt.Scanln(&text)

		if text == "exit" {
			break
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
	wg.Done()

	// register the grpc services
	proto.RegisterDistributedServiceServer(grpcDServer, dserver)
	proto.RegisterAuctionServiceServer(grpcDServer, dserver)

	serveError := grpcDServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
	log.Printf("Started gRPC service")
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

/*----  Published services for client  TO DO----*/
func (ds *DServer) AskForMaster(ctx context.Context, in *proto.Empty) (*proto.Peer, error) {
	// return the current master
	return &proto.Peer{
		Address: masterAddr,
		Port:    int32(masterPort),
		Id:      int32(masterId)}, nil
}

func (ds *DServer) Bid(ctx context.Context, in *proto.Amount) (*proto.Empty, error) {
	if !iAmMaster(ds) {
		return nil, status.Errorf(codes.FailedPrecondition, "I'am not the master")
	}
	// i'am the master
	// TO DO
	return &proto.Empty{}, nil
}

func (ds *DServer) Result(ctx context.Context, in *proto.Empty) (*proto.Outcome, error) {
	if !iAmMaster(ds) {
		return nil, status.Errorf(codes.FailedPrecondition, "I'am not the master")
	}
	// i'am the master
	// TO DO
	return &proto.Outcome{
		Highest: 100,
		Closed:  false}, nil
}

/*------ Published services for other peers ----*/
func (ds *DServer) Election(ctx context.Context, in *proto.Peer) (*proto.Empty, error) {
	if ds.id < int(in.Id) {
		return nil, status.Errorf(codes.FailedPrecondition, "Peer %d has a lower ID than peer %d", ds.id, in.Id)
	}

	log.Printf("Peer [%s:%d, %d] received election message from [%s:%d, %d]\n", ds.address, ds.port, ds.id, in.Address, in.Port, in.Id)

	// send election to bigger ID than me and reply with ack
	findMaster(ds)
	return &proto.Empty{}, nil
}

func (ds *DServer) Coordinator(ctx context.Context, in *proto.Peer) (*proto.Empty, error) {
	select {
	case coordination_msg <- true:
		break
	}

	log.Printf("Peer [%s:%d, %d] is the new coordinator.\n", in.Address, in.Port, in.Id)
	masterAddr = in.Address
	masterPort = int(in.Port)
	masterId = int(in.Id)

	return &proto.Empty{}, nil
}

func (ds *DServer) Ping(ctx context.Context, in *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, nil
}

/*--- Function to make code more readable and manage the master ---*/
func iAmMaster(ds *DServer) bool {
	return (ds.address == masterAddr && ds.port == masterPort && ds.id == masterId)
}

func findMaster(ds *DServer) {

	me := &proto.Peer{
		Address: ds.address,
		Port:    int32(ds.port),
		Id:      int32(ds.id),
	}
	if !sendElectionToBigger(ds, me) {
		// i'am the master
		masterAddr = ds.address
		masterPort = ds.port
		masterId = ds.id
		log.Printf("I'am the current master!\n")
		sendCoordinatorToLower(ds, me)
	} else {
		// wait for coordination message
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-timer.C:
			log.Printf("No one of the bigger id peer send a coordination msg")
			// retry
			findMaster(ds)
			break
		case <-coordination_msg:
			break
		}
	}
}

// election message to bigger id peers (1 second to reply)
func sendElectionToBigger(ds *DServer, msg *proto.Peer) bool {
	found_bigger_active := false

	for index, conn := range peers {
		if index > ds.id {
			_, err := conn.Election(context.Background(), msg)
			if err != nil {
				//log.Printf("Failed to send election message to peer id %d", index)
				// continue with other peers
				continue
			}
			log.Printf("Sent selection message to peer id %d", index)
			found_bigger_active = true
		}
	}
	return found_bigger_active
}

func sendCoordinatorToLower(ds *DServer, msg *proto.Peer) {
	for index, conn := range peers {
		if index < ds.id {
			_, err := conn.Coordinator(context.Background(), msg)
			if err != nil {
				//log.Printf("Failed to send coordinator message to peer id %d", index)
				// continue with other peers
				continue
			}
			log.Printf("Sent coordinator message to peer id %d", index)
		}
	}
}

/* --- functions to monitor master status ---*/
func checkMaster(ds *DServer) {
	for {
		select {
		case <-time.NewTicker(5 * time.Second).C:
			if !iAmMaster(ds) {
				log.Print("Check if the master is available")
				if !pingMaster() {
					// master is not available
					log.Printf("Master is not available")
					findMaster(ds)
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
