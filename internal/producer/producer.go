// This contains all necessary tools for the producer to accept connections and process the recieved data
package producer

import (
	"fmt"
	"runtime"
	"time"

	"github.com/SIGBlockchain/project_aurum/internal/block"
)

func PrintMemstats(ms runtime.MemStats) {
	// useful commands: go run -gcflags='-m -m' main.go <main flags>
	fmt.Printf("Bytes of allocated heap objects: %d", ms.Alloc)
	fmt.Printf("Cumulative bytes allocated for heap objects: %d", ms.TotalAlloc)
	fmt.Printf("Count of heap objects allocated: %d", ms.Mallocs)
	fmt.Printf("Count of heap objects freed: %d", ms.Frees)
}

func TriggerInterval(intervalChannel chan bool, productionInterval time.Duration) {
	// Triggers block production case
	time.Sleep(productionInterval)
	intervalChannel <- true
}

func CalculateInterval(youngestBlockHeader block.BlockHeader, productionInterval time.Duration, intervalChannel chan bool) {
	var lastTimestamp = time.Unix(0, youngestBlockHeader.Timestamp)
	timeSince := time.Since(lastTimestamp)
	if timeSince.Nanoseconds() >= productionInterval.Nanoseconds() {
		go TriggerInterval(intervalChannel, time.Duration(0))
	} else {
		diff := productionInterval.Nanoseconds() - timeSince.Nanoseconds()
		go TriggerInterval(intervalChannel, time.Duration(diff))
	}
}
