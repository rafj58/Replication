package main

import (
	"context"
	"flag"
	"log"
	"strconv"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name string
}

var (
	firstAddr  = flag.String("sAddr", "localhost", "server address") // retrive first line from conf File
	firstPort  = flag.Int("sPort", 1000, "server port")              // as before
	masterAddr = flag.String("sAddr", "localhost", "server address")
	masterPort = flag.Int("sPort", 1000, "server port")
	clientName = flag.String("cName", "Anonymous", "client name")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		name: *clientName,
	}

	// Connect
	connect(client)
}

func connect(client *Client) {
	// ask for the master
	first := connectToFirst()
	// update masterAddr and masterPort

	stream, err := first.AskForMaster(context.Background(), &proto.Empty{})
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Printf("The master is: %s:%d", stream.Address, stream.Port)
	}
	*masterAddr = stream.Address
	*masterPort = int(stream.Port)

	// connect to the master
	master := connectToMaster()
	// make a bid, it is a test
	master.Bid(context.Background(), &proto.Amount{
		Amount: 10,
	})
	// to do ...
}

func connectToMaster() proto.AuctionServiceClient {
	// Dial the server at the specified port.
	conn, err := grpc.Dial(*masterAddr+":"+strconv.Itoa(*masterPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *masterPort)
	} else {
		log.Printf("Connected to the master at port %d\n", *masterPort)
	}
	return proto.NewAuctionServiceClient(conn)
}

func connectToFirst() proto.DistributedServiceClient {
	// Dial the server at the specified port.
	conn, err := grpc.Dial(*firstAddr+":"+strconv.Itoa(*firstPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *firstPort)
	} else {
		log.Printf("Connected to the master at port %d\n", *firstPort)
	}
	return proto.NewDistributedServiceClient(conn)
}
