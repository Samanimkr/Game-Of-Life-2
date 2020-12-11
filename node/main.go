package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
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
	P               Params
	World           [][]byte
	NextAddress     string
	PreviousAddress string
	WorkerHeight    int
}

type Node struct{}

var PARAMS Params
var WORLD [][]byte
var PREVIOUS_ROW []byte
var NEXT_ROW []byte
var WORKER_HEIGHT int

var NEXT_NODE *rpc.Client
var PREVIOUS_NODE *rpc.Client

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

func worker(world [][]byte, workerOut chan<- byte) {
	// Create a temporary empty world
	tempWorld := createWorld(WORKER_HEIGHT+2, PARAMS.ImageWidth)

	// Loop through the worker's section of the world
	for y := 1; y <= WORKER_HEIGHT; y++ {
		for x := 0; x < PARAMS.ImageWidth; x++ {
			// Get number of alive neighbours
			numAliveNeighbours := aliveNeighbours(world, y, x, PARAMS)

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
	for y := 0; y < WORKER_HEIGHT; y++ {
		for x := 0; x < PARAMS.ImageWidth; x++ {
			workerOut <- tempWorld[y+1][x]
		}
	}
}

func (n *Node) GetPreviousRow(x int, reply *[]byte) (err error) {
	lastRowToSend := make([]byte, PARAMS.ImageWidth)
	for i := range lastRowToSend {
		lastRowToSend[i] = WORLD[WORKER_HEIGHT-1][i]
	}

	*reply = lastRowToSend

	return
}

func (n *Node) GetNextRow(x int, reply *[]byte) (err error) {
	firstRowToSend := make([]byte, PARAMS.ImageWidth)
	for i := range firstRowToSend {
		firstRowToSend[i] = WORLD[0][i]
	}

	*reply = firstRowToSend

	return
}

func (n *Node) SendData(args NodeArgs, x *int) (err error) {
	WORLD = args.World
	PARAMS = args.P
	WORKER_HEIGHT = args.WorkerHeight

	*x = 0
	return
}

func (n *Node) SendAddresses(args NodeArgs, x *int) (err error) {
	if NEXT_NODE == nil {
		nextNode, error := rpc.Dial("tcp", args.NextAddress)
		if error != nil {
			log.Fatal("Unable to connect1", error)
		}
		NEXT_NODE = nextNode
	}

	if PREVIOUS_NODE == nil {
		prevNode, error := rpc.Dial("tcp", args.PreviousAddress)
		if error != nil {
			log.Fatal("Unable to connect2", error)
		}
		PREVIOUS_NODE = prevNode
	}

	nextRowToReceive := make([]byte, args.P.ImageWidth)
	NEXT_NODE.Call("Node.GetNextRow", 0, &nextRowToReceive)
	NEXT_ROW = nextRowToReceive

	prevRowToReceive := make([]byte, args.P.ImageWidth)
	PREVIOUS_NODE.Call("Node.GetPreviousRow", 0, &prevRowToReceive)
	PREVIOUS_ROW = prevRowToReceive

	*x = 0

	return
}

func (n *Node) Start(x int, reply *[][]byte) (err error) {
	tempWorld := createWorld(WORKER_HEIGHT+2, PARAMS.ImageWidth)
	for i := range tempWorld {
		tempWorld[i] = make([]byte, PARAMS.ImageWidth)
	}

	for y := 0; y < WORKER_HEIGHT+2; y++ {
		for x := 0; x < PARAMS.ImageWidth; x++ {
			if y == 0 {
				tempWorld[y][x] = PREVIOUS_ROW[x]
			} else if y == WORKER_HEIGHT+1 {
				tempWorld[y][x] = NEXT_ROW[x]
			} else {
				tempWorld[y][x] = WORLD[y-1][x]
			}
		}
	}

	workerOut := make(chan byte)
	go worker(tempWorld, workerOut)

	newSplit := createWorld(WORKER_HEIGHT, PARAMS.ImageWidth)

	// Get all the updated cells from the 'workerOut' channel
	for y := 0; y < WORKER_HEIGHT; y++ {
		for x := 0; x < PARAMS.ImageWidth; x++ {
			// Get the new value
			newSplit[y][x] = <-workerOut

			// Update that cell in 'world'
			if WORLD[y][x] != newSplit[y][x] {
				WORLD[y][x] = newSplit[y][x]
			}
		}
	}

	*reply = WORLD
	return
}

// Kill function
func (n *Node) Kill(x int, reply *int) (err error) {
	NEXT_NODE.Close()
	PREVIOUS_NODE.Close()

	*reply = 0
	os.Exit(0)
	return
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread() // not sure what this does but was in skeleton

	portPtr := flag.String("port", ":8031", "listening on this port")
	flag.Parse()          // call after all flags are defined to parse command line into flags
	rpc.Register(&Node{}) // WHAT DOES THIS DO?

	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	// defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln) // accepts connections on ln and serves requests to server for each incoming connection
}
