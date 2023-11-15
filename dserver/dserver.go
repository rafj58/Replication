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
	// mutual exclusion on master variables (address & port)
	mutex sync.Mutex
}

var (
	my_row   = flag.Int("row", 1, "Indicate the row of parameter file for this peer") // set with "-row <port>" in terminal
	name     = flag.String("name", "peer", "name of the peer")
	confFile = "confFile.csv"
	// store  tcp connection to others peers of the distributed server -id is the key
	peers = make(map[int]proto.DistributedServiceClient)
	// wait for listen to be open
	wg sync.WaitGroup
	// master info
	masterAddr string
	masterPort int
	masterId   int
)

func main() {
	flag.Parse()

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
			fmt.Printf("Your settings are : %s address, %d port, %d id\n", peerAddress, peerPort, peerId)
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

	me := &proto.Peer{
		Address: dserver.address,
		Port:    int32(dserver.port),
		Id:      int32(dserver.id),
	}

	if !sendElectionToBigger(dserver, me) {
		// no one bigger replied, i'am the master
		masterAddr = dserver.address
		masterPort = dserver.port
		masterId = dserver.id
	}
}

func StartListen(dserver *DServer) {
	// Create a new grpc server
	grpcDServer := grpc.NewServer()

	// Make the peer listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", dserver.address, strconv.Itoa(dserver.port)))

	if err != nil {
		log.Fatalf("Could not create the peer %v", err)
	}
	log.Printf("Started peer receiving at address: %s and at port: %d\n", dserver.address, dserver.port)
	wg.Done()

	// Register the grpc services
	proto.RegisterDistributedServiceServer(grpcDServer, dserver)
	proto.RegisterAuctionServiceServer(grpcDServer, dserver)

	serveError := grpcDServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
	log.Printf("Started gRPC service")
}

// Connect to other peer
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

/*----  Published services for client ----*/
func (ds *DServer) AskForMaster(ctx context.Context, in *proto.Empty) (*proto.Peer, error) {
	// return the current master
	return &proto.Peer{
		Address: masterAddr,
		Port:    int32(masterPort)}, nil
}

func (ds *DServer) Bid(ctx context.Context, in *proto.Amount) (*proto.Acknowledge, error) {
	if !iAmMaster(ds) {
		return nil, status.Errorf(codes.FailedPrecondition, "I'am not the master")
	}
	// i'am the master
	// change the amount TO DO
	return &proto.Acknowledge{Status: true}, nil
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

func (ds *DServer) Election(ctx context.Context, in *proto.Peer) (*proto.Acknowledge, error) {
	if ds.id < int(in.Id) {
		return nil, status.Errorf(codes.FailedPrecondition, "Peer %d has a lower ID than peer %d", ds.id, in.Id)
	}

	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	log.Printf("Peer [%s:%d, %d] received election message from [%s:%d, %d]\n",
		ds.address, ds.port, ds.id, in.Address, in.Port, in.Id)

	// forward election to bigger ID than me and reply with ack
	if !sendElectionToBigger(ds, in) {
		log.Fatal("Some error!")
	} else {
		// no one bigger replied, i'am the coordinator
		// TO DO
	}
	// to remove
	return &proto.Acknowledge{
		Status: true}, nil
}

func (ds *DServer) Coordinator(ctx context.Context, in *proto.Peer) (*proto.Acknowledge, error) {
	// i am the bully
	log.Printf("Peer [%s:%d, %d] is the new coordinator.\n", ds.address, ds.port, ds.id)
	masterAddr = ds.address
	masterPort = ds.port
	return &proto.Acknowledge{
		Status: true}, nil
}

func setMaster(ds *DServer) {
	// set election message to peer with id bigger than mine
	found_bigger := false

	if !found_bigger {
		log.Fatal("Some error!")
	}
}

func iAmMaster(ds *DServer) bool {
	return (ds.address == masterAddr && ds.port == masterPort && ds.id == masterId)
}

func sendElectionToBigger(ds *DServer, msg *proto.Peer) bool {
	// foward election to bigger id peers
	timeout := 1 * time.Second
	found_bigger := false
	for index, conn := range peers {
		if index > ds.id {
			select {
			case <-time.After(timeout):
				log.Printf("Timeout while trying to send election message to peer id: %d", index)
				// continue with other peers
				continue
			default:
				_, err := conn.Election(context.Background(), msg)
				if err != nil {
					log.Printf("Failed to send election message to peer id %d: %v", index, err)
					// continue with other peers
					continue
				}
				// check resp
				found_bigger = true
			}
		}
	}
	return found_bigger
}
