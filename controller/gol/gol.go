package gol

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/util"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type Args struct {
	P     Params
	World [][]byte
	Turn  int
}

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

func getAliveCells(p Params, world [][]byte) []util.Cell {
	currentAliveCells := make([]util.Cell, 0) // create aliveCells slice
	for y := 0; y < p.ImageHeight; y++ {      // go through all cells in world
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] != 0 {
				currentAliveCells = append(currentAliveCells, util.Cell{X: x, Y: y}) // adds current cell to the aliveCells slice
			}
		}
	}
	return currentAliveCells
}

func outputPGM(world [][]byte, c distributorChannels, p Params, turn int) {
	c.ioCommand <- ioCommand(ioOutput)
	outputFileName := fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
	c.ioFilename <- outputFileName
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- ImageOutputComplete{turn, outputFileName}
}

func engine(p Params, c distributorChannels, keyPresses <-chan rune) {
	// create slice to store world
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	c.ioCommand <- ioCommand(ioInput)                             // send read command down command channel
	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth) // gets file name from putting file dimensions together
	c.ioFilename <- filename                                      // sends file name to the fileName channel

	// populate world
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
	}

	// connect to engine
	controller := engineConnection()

	var args Args
	response := new([][]byte)

	for turns := 0; turns < p.Turns; turns++ {
		request := Args{
			World: world,
			Turn:  turns,
			P:     p,
		}

		controller.Call("Engine.Start", request, &response)
		args.World = *response
	}
	outputPGM(*response, c, p, p.Turns)
}

func engineConnection() *rpc.Client {
	// connect to engine
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	controller, error := rpc.Dial("tcp", *server)
	if error != nil {
		log.Fatal("Unable to connect", error)
	}

	return controller
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	ioFileName := make(chan string)
	ioOutput := make(chan uint8)
	ioInput := make(chan uint8)

	distributorChannels := distributorChannels{
		events,
		ioCommand,
		ioIdle,
		ioFileName,
		ioOutput,
		ioInput,
		keyPresses,
	}

	go engine(p, distributorChannels, keyPresses)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFileName,
		output:   ioOutput,
		input:    ioInput,
	}
	go startIo(p, ioChannels)
}
