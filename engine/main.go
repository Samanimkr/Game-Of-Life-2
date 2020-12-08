package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"runtime"
)

// Args COMMENT
type Args struct {
	P     Params
	World [][]byte
}

type Engine struct{}

type AliveCellsReply struct {
	AliveCells     int
	CompletedTurns int
}

type SaveReply struct {
	CompletedTurns int
	World          [][]byte
}

type PauseReply struct {
	CompletedTurns int
	World          [][]byte
}

type IsAlreadyRunningReply struct {
	P                Params
	World            [][]byte
	IsAlreadyRunning bool
}

var WORLD [][]byte
var PARAMS Params
var ALIVECELLS int
var COMPLETEDTURNS int
var PAUSECHANNEL = make(chan bool, 1)
var FINISHCHANNEL = make(chan bool, 1)

// IsAlreadyRunning function
func (e *Engine) IsAlreadyRunning(p Params, reply *IsAlreadyRunningReply) (err error) {
	fmt.Println("p == PARAMS |", p == PARAMS)
	fmt.Println("WORLD != nil |", WORLD != nil)
	fmt.Println("COMPLETEDTURNS-1 > 0 |", COMPLETEDTURNS-1 > 0)
	fmt.Println("COMPLETEDTURNS", COMPLETEDTURNS)

	if p == PARAMS && WORLD != nil && COMPLETEDTURNS-1 > 0 {
		*reply = IsAlreadyRunningReply{
			IsAlreadyRunning: true,
			P:                PARAMS,
			World:            WORLD,
		}

		return
	}
	WORLD = nil
	PARAMS = p
	COMPLETEDTURNS = 0

	*reply = IsAlreadyRunningReply{
		IsAlreadyRunning: false,
	}
	return
}

// Start function
func (e *Engine) Start(args Args, reply *[][]byte) (err error) {
	PARAMS = args.P
	fmt.Println("Start 1")
	fmt.Println("Start 1 args.P: ", args.P)
	fmt.Println("Start 1 args.World: ", args.World != nil)
	WORLD = distributor(args.P, args.World)
	fmt.Println("\nStart 2 WORLD: ", WORLD)
	*reply = WORLD

	return
}

// Continue function
func (e *Engine) Continue(args Args, reply *[][]byte) (err error) {
	fmt.Println("000")

	asd := <-FINISHCHANNEL
	fmt.Println("Finish: ", asd)
	*reply = WORLD

	return
}

// Save function
func (e *Engine) Save(x int, reply *SaveReply) (err error) {
	saveReply := SaveReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = saveReply

	return
}

// Pause function
func (e *Engine) Pause(x int, reply *PauseReply) (err error) {
	PAUSECHANNEL <- true
	pauseReply := PauseReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = pauseReply

	return
}

// Execute function
func (e *Engine) Execute(x int, reply *PauseReply) (err error) {
	PAUSECHANNEL <- false
	executeReply := PauseReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = executeReply

	return
}

// Quit funtion
func (e *Engine) Quit(x int, reply *int) (err error) {
	*reply = COMPLETEDTURNS

	return
}

// GetAliveCells ...
func (e *Engine) GetAliveCells(x int, reply *AliveCellsReply) (err error) {
	aliveCells := AliveCellsReply{
		AliveCells:     ALIVECELLS,
		CompletedTurns: COMPLETEDTURNS,
	}
	*reply = aliveCells

	return
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread() // not sure what this does but was in skeleton

	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "listening on this port")
	flag.Parse()                             // call after all flags are defined to parse command line into flags
	rpc.Register(&Engine{})                  // WHAT DOES THIS DO?
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection
}
