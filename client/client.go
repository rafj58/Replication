package main

import (
	"context"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"strconv"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name string
}

var (
	clientName = flag.String("cName", "Anonymous", "client name")
)

const (
	configFilePath = "../confFile.csv"
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	// client := &Client{
	// 	name: *clientName,
	// }

	// Read Server addresses and ports from file
	addrPortIds := GetServerAddrPortIdFromFile()

	// Get master and connect to it
	AskAndConnectToMaster(addrPortIds)

}

func GetServerAddrPortIdFromFile() [][]string {
	file, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf(err.Error())
	} else {
		log.Printf("Successfully opened config file for reading")
	}
	reader := csv.NewReader(file)
	addrPortIds, err := reader.ReadAll()
	if err != nil {
		log.Fatalf(err.Error())
	} else {
		log.Printf("Succesfully read config file")
	}
	return addrPortIds
}

func AskAndConnectToMaster(addrPortIds [][]string) proto.AuctionServiceClient {
	var serviceClient proto.AuctionServiceClient
	for _, addrPortId := range addrPortIds {
		service, err := ConnectToServerNode(addrPortId[0], addrPortId[1])
		if err != nil {
			continue
		}
		masterNode, err := service.AskForMaster(context.Background(), &proto.Empty{})
		if err != nil {
			log.Printf("Attempt to communicate with server node at %s:%s resulted in error: %s", addrPortId[0], addrPortId[1], err)
			continue
		} else {
			log.Printf("Successfully recieved address and port for master node from node at %s:%s", addrPortId[0], addrPortId[1])
		}
		serviceClient, err = ConnectToServerNode(masterNode.GetAddress(), strconv.Itoa(int(masterNode.GetPort())))
		if err != nil {
			log.Fatalf("Attempt to communicate with master node at %s:%d resulted in error: %s", masterNode.GetAddress(), masterNode.GetPort(), err)
		}
		break
	}
	if serviceClient == nil {
		log.Fatalf("Could not connect to any server node")
	}
	return serviceClient
}

func ConnectToServerNode(address string, port string) (proto.AuctionServiceClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial(address+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect to servernode at %s:%s", address, port)
	} else {
		log.Printf("Connected to servernode at %s:%s", address, port)
	}
	return proto.NewAuctionServiceClient(conn), err
}
