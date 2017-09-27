package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	nThrow := flag.Int("n", 1e9, "number of throws")
	nCPU := flag.Int("c", runtime.NumCPU(), "number of CPUs to use")
	flag.Parse()
	runtime.GOMAXPROCS(*nCPU) // Set number of OS threads to use.
	parts := make(chan int)   // Channel to collect partial results.

	count := 0
	mCount := sync.RWMutex{}

	done := make(chan int)

	// Kick off parallel tasks.
	for i := 0; i < *nCPU; i++ {
		go func(me int) {
			// Create local PRNG to avoid contention.
			r := rand.New(rand.NewSource(int64(me)))
			n := *nThrow / *nCPU
			hits := 0

			c := 0
			// Do the throws.
			for i := 0; i < n; i++ {
				x := r.Float64()
				y := r.Float64()
				if x*x+y*y < 1 {
					hits++
				}

				// increase count for print process
				c++
				if c == 1000 || i == n-1 {
					mCount.Lock()
					count += c
					mCount.Unlock()
					c = 0
				}
			}
			parts <- hits // Send the result back.
		}(i)
	}

	// print the process of calculation
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				mCount.RLock()
				fmt.Printf("\r%.2f%%", float32(count)/float32(*nThrow)*100)
				mCount.RUnlock()

				// print every 1 second
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Aggregate partial results.
	hits := 0
	for i := 0; i < *nCPU; i++ {
		hits += <-parts
	}
	close(done)

	pi := 4 * float64(hits) / float64(*nThrow)
	fmt.Printf("\rPI = %f\n", pi)
	fmt.Printf("time cost: %f s\n", time.Since(start).Seconds())
}
