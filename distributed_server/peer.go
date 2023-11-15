package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "MutualExclusion/grpc"
)

// link for Ricart & Agrawala algorithm https://www.geeksforgeeks.org/ricart-agrawala-algorithm-in-mutual-exclusion-in-distributed-system/
// we need only the port where we receive messages
// output port is decided automatically randomly by operating system

type Peer struct {
	proto.UnimplementedMutualExlusionServiceServer
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
	my_row = flag.Int("row", 1, "Indicate the row of parameter file for this peer") // set with "-row <port>" in terminal
	name   = flag.String("name", "peer", "name of the peer")
	// Lamport variable
	lamport_time = 0
	confFile     = "confFile.csv"
	// default values for address and port
	my_address = "127.0.0.1"
	my_port    = 50050
	// store tcp connection to others peers
	peers = make(map[string]proto.MutualExlusionServiceClient)
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
	}

	if !found {
		fmt.Printf("Row with parameters not founded\n")
		return
	}

	peer := &Peer{
		name:    *name,
		address: my_address,
		port:    my_port,
	}
	// wait for opening port to listen
	wg.Add(1)

	// open the port to new connections
	go StartListen(peer)

	wg.Wait()
	// Preparate tcp connection to the others client
	connectToOthersPeer(peer)

	// user interface menu
	doSomething()
}

func StartListen(peer *Peer) {
	// Create a new grpc server
	grpcPeer := grpc.NewServer()

	increaseTime()
	// Make the peer listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", peer.address, strconv.Itoa(peer.port)))

	if err != nil {
		log.Fatalf("Could not create the peer %v", err)
	}
	log.Printf("Lamport %d: Started peer receiving at address: %s and at port: %d\n", lamport_time, peer.address, peer.port)
	wg.Done()

	// Register the grpc service
	increaseTime()
	proto.RegisterMutualExlusionServiceServer(grpcPeer, peer)
	serveError := grpcPeer.Serve(listener)

	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
	log.Printf("Lamport %d: Started gRPC service", lamport_time)

}

// Connect to others peer
func connectToOthersPeer(p *Peer) {
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

func connectToPeer(address string, port int) proto.MutualExlusionServiceClient {
	// Dial doesn't check if the peer at that address:host is effectivly on (simply prepare TCP connection)
	increaseTime()
	conn, err := grpc.Dial(address+":"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Lamport %d: Could not connect to peer %s at port %d", lamport_time, address, port)
	} else {
		log.Printf("Lamport %d: Created TCP connection to the %s address at port %d\n", lamport_time, address, port)
	}
	return proto.NewMutualExlusionServiceClient(conn)
}

func (peer *Peer) AskPermission(ctx context.Context, in *proto.Question) (*proto.Answer, error) {
	setTime(int(in.Time))
	// check if the peer requesting permission is not in the list of connected peers
	// it can be a reconnected peer or one not present in the configuration file
	peerRef := in.ClientReference.ClientAddress + ":" + strconv.Itoa(int(in.ClientReference.ClientPort))
	log.Printf("Lamport %d: Peer [%s] asked for a mutual exection", lamport_time, peerRef)
	found := false
	for index := range peers {
		if index == peerRef {
			found = true
			break
		}
	}
	// receive request from a not known peer
	if !found {
		connection := connectToPeer(in.ClientReference.ClientAddress, int(in.ClientReference.ClientPort))
		peers[peerRef] = connection
	}
	// Ricart–Agrawala Algorithm
	if (state == Held) || (state == Wanted && (in.Time > int32(myRequestTime))) {
		// queue the reply (just wait unline i'm done)
		for state == Held || state == Wanted {
			time.Sleep(500 * time.Millisecond)
		}
	}
	log.Printf("Lamport %d: Peer [%s] authorized to do mutual exection", lamport_time, peerRef)
	increaseTime()
	return &proto.Answer{
		Reply: true,
		Time:  int32(lamport_time),
	}, nil

}

func doSomething() {
	for {
		var text string
		log.Printf("Insert 'mutual' to do mutual execution or 'exit' to quit or anything else "+
			"to increment time [Actual Lamport Time: %d] ", lamport_time)
		fmt.Scanln(&text)

		increaseTime() // an event occurred
		if text == "exit" {
			break
		}

		if text != "mutual" {
			continue
		}

		state = Wanted
		myRequestTime = lamport_time

		// Peers enters the critical section if it has received the REPLY message from all other sites.
		peerRef := &proto.ClientReference{
			ClientAddress: my_address,
			ClientPort:    int32(my_port),
			ClientName:    *name,
		}
		for index, peer := range peers {
			increaseTime()
			log.Printf("Lamport %d: Asked Peer [%s] for permission", lamport_time, index)
			answer, err := peer.AskPermission(context.Background(),
				&proto.Question{
					ClientReference: peerRef,
					Time:            int32(myRequestTime),
				})
			if err != nil {
				log.Printf("Lamport %d: Peer [%s] no more available, removed from connected peers", lamport_time, index)
				delete(peers, index)
				continue
			} else {
				setTime(int(answer.Time))
				log.Printf("Lamport %d: Got permission from peer [%s]", lamport_time, index)
			}
		}
		// do critical section
		criticalSection()
	}
}

func criticalSection() {
	increaseTime()
	state = Held
	log.Printf("Lamport %d: Starting critical section", lamport_time)
	time.Sleep(time.Duration(rand.Intn(4)+10) * time.Second)
	increaseTime()
	log.Printf("Lamport %d: Ending critical section", lamport_time)
	state = Released
}

func increaseTime() {
	lamport_time++
}

func setTime(received int) {
	max := math.Max(float64(received), float64(lamport_time))
	lamport_time = int(max + 1)
}
