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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "Replication/grpc"
)

type DServer struct {
	proto.UnimplementedAuctionServiceServer
	proto.UnimplementedDistributedServiceServer
	name    string
	address string
	port    int
}

var (
	my_row   = flag.Int("row", 1, "Indicate the row of parameter file for this peer") // set with "-row <port>" in terminal
	name     = flag.String("name", "peer", "name of the peer")
	confFile = "confFile.csv"
	// store tcp connection to others peers of the distributed server
	peers = make(map[string]proto.DistributedServiceClient)
	// wait for listen to be open
	wg sync.WaitGroup
	// master info
	masterAddr string
	masterPort int
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
		peerRef := row[0] + ":" + row[1]

		if index == *my_row {
			fmt.Printf("Your settings are : %s address, %d port\n", peerAddress, peerPort)
			dserver = &DServer{
				name:    *name,
				address: peerAddress,
				port:    peerPort,
			}
			found = true
		} else {
			// retrieve connection
			connection := connectToPeer(peerAddress, peerPort)
			// add to map
			peers[peerRef] = connection
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

	// user interface menu
	chooseMaster()
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

// BUlly algorithm to choose the master
func chooseMaster() {

}

// Published services
func (ds *DServer) AskForMaster(ctx context.Context, in *proto.Empty) (*proto.Master, error) {
	// return the current master
	return &proto.Master{
		Address: masterAddr,
		Port:    int32(masterPort)}, nil
}

func (ds *DServer) Bid(ctx context.Context, in *proto.Amount) (*proto.Acknowledge, error) {
	// TO DO
}

func (ds *DServer) Result(ctx context.Context, in *proto.Empty) (*proto.Outcome, error) {
	// TO DO
}
