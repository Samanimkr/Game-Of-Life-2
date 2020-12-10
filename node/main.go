package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
)

type Node struct{}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	portPtr := flag.String("port", ":8031", "listening on this port")
	flag.Parse()                             // call after all flags are defined to parse command line into flags
	rpc.Register(&Node{})                    // WHAT DOES THIS DO?
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection

}
