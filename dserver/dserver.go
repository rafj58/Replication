package main

import (
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

// link for Ricart & Agrawala algorithm https://www.geeksforgeeks.org/ricart-agrawala-algorithm-in-mutual-exclusion-in-distributed-system/
// we need only the port where we receive messages
// output port is decided automatically randomly by operating system

type DServer struct {
	proto.UnimplementedAuctionServiceServer
	proto.UnimplementedDistributedServiceServer
	name    string
	address string
	port    int
}

// peer states
const (
	Released int = 0
	Wanted       = 1
	Held         = 2
)

var (
	my_row   = flag.Int("row", 1, "Indicate the row of parameter file for this peer") // set with "-row <port>" in terminal
	name     = flag.String("name", "peer", "name of the peer")
	confFile = "../confFile.csv"
	// default values for address and port
	my_address = "127.0.0.1"
	my_port    = 50050
	// store tcp connection to others peers
	peers = make(map[string]proto.UnimplementedDistributedServiceServer)
	// state of the distributed mutex
	state = Released
	// lamport time of this peers request
	myRequestTime = 0
	// wait for listen before try to connect to other peers
	wg sync.WaitGroup
)

func main() {
	flag.Parse()

	// read from confFile.txt and set the peer values
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

	found := false
	for index, row := range rows {
		if index == *my_row {
			fmt.Printf("Your settings are : %s address, %s port\n", row[0], row[1])
			my_address = row[0]
			my_port, _ = strconv.Atoi(row[1])
			found = true
			break
		}
		dserver := &DServer{
			name:    *name,
			address: my_address,
			port:    my_port,
		}
		connectToNode(dserver)
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
	// Preparate tcp connection to the others client
	connectToOthersPeer(dserver)

	// user interface menu
	doSomething()
}

func StartListen(dserver *DServer) {
	// Create a new grpc server
	grpcDServer := grpc.NewServer()

	increaseTime()
	// Make the peer listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", dserver.address, strconv.Itoa(dserver.port)))

	if err != nil {
		log.Fatalf("Could not create the peer %v", err)
	}
	log.Printf("Lamport %d: Started peer receiving at address: %s and at port: %d\n", lamport_time, dserver.address, dserver.port)
	wg.Done()

	// Register the grpc service
	increaseTime()
	proto.RegisterDistributedServiceServer(grpcDServer, dserver)
	serveError := grpcDServer.Serve(listener)

	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
	log.Printf("Lamport %d: Started gRPC service", lamport_time)

}

// Connect to others peer
func connectToNode(dserver *DServer) {
	// read csv file
	file, err := os.Open(confFile)
	if err != nil {
		log.Fatalf("Failed to open configuration file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read file data: %v", err)
	}

	// try to connect to other peers
	for index, row := range rows {
		if len(row) < 2 || (index == *my_row) {
			// ignore corrupted rows and me
			continue
		}
		peerAddress := row[0]
		peerPort, _ := strconv.Atoi(row[1])
		peerRef := row[0] + ":" + row[1]
		// retrieve connection
		connection := connectToPeer(peerAddress, peerPort)
		// add to map
		peers[peerRef] = connection
	}
}

func connectToPeer(address string, port int) proto.DistributedServiceClient {
	// Dial doesn't check if the peer at that address:host is effectivly on (simply prepare TCP connection)
	increaseTime()
	conn, err := grpc.Dial(address+":"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Lamport %d: Could not connect to peer %s at port %d", lamport_time, address, port)
	} else {
		log.Printf("Lamport %d: Created TCP connection to the %s address at port %d\n", lamport_time, address, port)
	}
	return proto.NewDistributedServiceClient(conn)
}
