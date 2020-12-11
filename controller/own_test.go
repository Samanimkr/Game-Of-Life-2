package main

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func Benchmark(b *testing.B) {

	dimensions := []gol.Params{
		// {ImageWidth: 16, ImageHeight: 16},
		// {ImageWidth: 128, ImageHeight: 128},
		{ImageWidth: 512, ImageHeight: 512},
	}
	for _, params := range dimensions {
		for i := 0; i < 3; i++ {
			params.Threads = 16
			params.Turns = 100
			os.Stdout = nil
			testName := fmt.Sprintf("Testing Benchmark For: %dx%d Turns: %d  Threads: %d ", params.ImageHeight, params.ImageWidth, params.Turns, params.Threads)
			b.Run(testName, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					events := make(chan gol.Event, 1000)
					keypresses := make(chan rune, 10)
					gol.Run(params, events, keypresses)
					for range events {
					}

				}
			})
		}
	}
}
