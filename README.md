# Replication
To run the code, make sure that the confFile.csv file has at least a number of rows equal or bigger to the number of peers server you want to create.
Each row in the file needs an ip-address, a port and an ID seperated by a comma.

The following command is to be understood if you are located with the terminal inside the main project folder.
To run the distributed server peer (dserver.go file) you need to provide, with -row, the line number of the configuration file to be assigned to the peer and optionally a name for the peer with -name. The rows start at 0.

```go run ./dserver/dserver.go -row 1```

When at least one of the server peers is running, you can create as many clients as you want.

```go run ./client/client.go```

If a peer with an ID greater than that of the server is created, it will become the new server.

**Server peer commands:**
- 'close' to terminate the auction (only the master can use this command)
- 'exit' to terminate the peer (if the master fails, one of the other peers will take its place)

**Client commands:**
- 'bid \<amount\>' to make a bid
- 'result' to see the current state of the auction
- 'exit' to terminate the client
