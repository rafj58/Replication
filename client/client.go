package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	proto "Replication/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	configFilePath = "confFile.csv"
)

var (
	masterAddr  string
	masterPort  int32
	serverNodes [][]string
)

func main() {
	// Read Server addresses and ports from file
	serverNodes = GetServerAddrPortIdFromFile()

	// Get master and connect to it
	auctionService := AskAndConnectToMaster(serverNodes)

	CommunicateWithService(auctionService)
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
			// the peer is not available, ask for the master to the next one
			continue
		}
		masterNode, err := service.AskForMaster(context.Background(), &proto.Empty{})
		if err != nil {
			log.Printf("Impossible to communicate with server node at %s:%s", addrPortId[0], addrPortId[1])
			continue
		} else {
			log.Printf("Successfully recieved address and port of the master node from node at %s:%s", addrPortId[0], addrPortId[1])
		}
		serviceClient, err = ConnectToServerNode(masterNode.GetAddress(), strconv.Itoa(int(masterNode.GetPort())))
		if err != nil {
			log.Printf("Impossible to communicate with master node at %s:%d", masterNode.GetAddress(), masterNode.GetPort())
		}
		masterAddr = masterNode.GetAddress()
		masterPort = masterNode.GetPort()
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

func CommunicateWithService(service proto.AuctionServiceClient) {
	var command string

	for {
		log.Printf("Type 'bid' followed by bid-amount to bid in the auction, 'result' to see the current \n" +
			"status of the auction or 'exit' to exit program")
		fmt.Scan(&command)

		switch command {
		case "bid":
			var amountstr string
			fmt.Scan(&amountstr)
			amount, err := strconv.Atoi(amountstr)

			if err != nil {
				log.Printf("Could not convert string to int. Error: %s", err.Error())
				continue
			}

			bid := proto.Amount{
				Amount: int32(amount),
			}
			_, err = service.Bid(context.Background(), &bid)

			if err != nil {
				if strings.Contains(err.Error(), "code = InvalidArgument") || strings.Contains(err.Error(), "code = DeadlineExceeded") {
					log.Println(err.Error())
					continue
				}
				log.Printf("Communication with master server %s:%d failed, searching for the new one", masterAddr, masterPort)
				AskAndConnectToMaster(serverNodes)
				continue
			}

			log.Printf("Bidded with amount : %s", strconv.Itoa(amount))

		case "result":
			response, err := service.Result(context.Background(), &proto.Empty{})
			if err != nil {
				log.Printf("Communication with server at %s:%d resulted in error: %s", masterAddr, masterPort, err.Error())
				AskAndConnectToMaster(serverNodes)
				continue
			}

			if response.Closed {
				log.Printf("The highest bid is %d and the auction has concluded", response.GetHighest())
			} else {
				log.Printf("The highest bid is %d and the auction is still ongoing", response.GetHighest())
			}

		case "exit":
			log.Printf("Exiting...")
			return

		default:
			log.Printf("Invalid command entered: %s", command)
		}
	}
}
