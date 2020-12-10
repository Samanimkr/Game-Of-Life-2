package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Args struct {
	P     Params
	World [][]byte
}

type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type NodeArgs struct {
	P            Params
	World        [][]byte
	NextAddress  string
	WorkerHeight int
}

type Node struct{}

var PARAMS Params
var WORLD [][]byte
var PREVIOUS_ROW []byte
var NEXT_ROW []byte
var WORKER_HEIGHT int

func mod(x, m int) int {
	return (x + m) % m
}

// Creates a 2D slice of the world depending on inputted height and width
func createWorld(height, width int) [][]byte {
	world := make([][]byte, height)
	for i := range world {
		world[i] = make([]byte, width)
	}
	return world
}

// Calculate the number of alive neighbours around the cell
func aliveNeighbours(world [][]byte, y, x int, p Params) int {
	neighbours := 0
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] != 0 {
					neighbours++
				}

			}
		}
	}
	return neighbours
}

func worker(world [][]byte, p Params, turn int, workerOut chan<- byte, workerHeight int) {
	// Create a temporary empty world
	tempWorld := createWorld(p.ImageHeight+2, p.ImageWidth)

	// Loop through the worker's section of the world
	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			// Get number of alive neighbours
			numAliveNeighbours := aliveNeighbours(world, y, x, p)

			// Calculate what's the new state of the cell depending on alive neighbours
			if world[y][x] == 255 {
				if numAliveNeighbours == 2 || numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
				}
			} else {
				if numAliveNeighbours == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
				}
			}
		}
	}

	// Send the updated world down the 'workerOut' channel
	for y := 0; y < workerHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			workerOut <- tempWorld[y+1][x]
		}
	}
}

func (n *Node) GetEndRow(prevRow []byte, reply *[]byte) (err error) {
	PREVIOUS_ROW = prevRow

	firstRowToSend := make([]byte, PARAMS.ImageWidth)
	for i := range firstRowToSend {
		firstRowToSend[i] = WORLD[0][i]
	}

	fmt.Println("GetEndRow! ", firstRowToSend)

	*reply = firstRowToSend

	return
}

func (n *Node) Start(args NodeArgs, reply *[][]byte) (err error) {
	fmt.Println("World Received")
	WORLD = args.World
	PARAMS = args.P

	nextNode, error := rpc.Dial("tcp", args.NextAddress)
	if error != nil {
		log.Fatal("Unable to connect", error)
	}

	lastRowToSend := make([]byte, args.P.ImageWidth)
	for i := range lastRowToSend {
		lastRowToSend[i] = WORLD[args.WorkerHeight-1][i]
	}

	nextRowToReceive := make([]byte, args.P.ImageWidth)
	nextNode.Call("Node.GetEndRow", lastRowToSend, &nextRowToReceive)
	NEXT_ROW = nextRowToReceive

	fmt.Println("NextRow: ", NEXT_ROW)

	*reply = make([][]byte, args.P.ImageWidth)
	return
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	portPtr := flag.String("port", ":8031", "listening on this port")
	flag.Parse()          // call after all flags are defined to parse command line into flags
	rpc.Register(&Node{}) // WHAT DOES THIS DO?
	fmt.Println("portPtr: ", *portPtr)
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection

}
