package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	goll "github.com/fabiofenoglio/goll"
)

func main() {
	// create an instance of LoadLimiter
	// accepting a max of 1000 over 20 seconds
	newLimiter, _ := goll.New(&goll.Config{
		MaxLoad:    1000,
		WindowSize: 20 * time.Second,
	})

	// not caring for multi-tenancy right now,
	// so let's switch to a single-tenant interface
	limiter := newLimiter.ForTenant("test")

	// we'll gather some stats to see how our boy performs
	startedAt := time.Now()
	requested := uint64(0)
	accepted := uint64(0)

	for i := 0; i < 100; i++ {
		// require a random amount of load from 10 to 50
		// to simulate various kind of requests
		requestedLoad := uint64(rand.Intn(50-10) + 10)

		requested += requestedLoad

		// submit the request to the limiter
		// SubmitUntil will retry automatically
		// for up to 10 seconds before giving up
		err := limiter.SubmitUntil(requestedLoad, 10*time.Second)
		if err != nil {
			// in case of timeout the returned error
			// will be of type: goll.ErrLoadRequestTimeout
			if errors.Is(err, goll.ErrLoadRequestTimeout) {
				fmt.Printf("request for load of %v was rejected and timed out while retrying (%v)\n", requestedLoad, err.Error())
			} else {
				fmt.Printf("request for load of %v failed (%v)\n", requestedLoad, err.Error())
			}
		} else {
			accepted += requestedLoad
			fmt.Printf("request for load of %v was accepted\n", requestedLoad)
		}

		// print a recap of the limiter status
		displayLimiterStatus(limiter)

		// sleep for a random amount of time from 0 to 1000ms
		// to simulate random requests pattern
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}

	// check if the limiter really did limit.
	// we expect a "total accepted" around 50/sec (1000 over 20 seconds)
	demoDuration := time.Now().Unix() - startedAt.Unix()
	fmt.Printf("**********************************\n")
	fmt.Printf("total duration: %d sec\n", demoDuration)
	fmt.Printf("total accepted: %.2v ( %.2f/sec )\n", accepted, float64(accepted)/float64(demoDuration))

}

// helper to display the status of the limiter.
func displayLimiterStatus(instance goll.SingleTenantStandaloneLoadLimiter) {
	stats, _ := instance.Stats()

	segmentsDesc := ""
	if len(stats.WindowSegments) > 0 {
		for _, segment := range stats.WindowSegments {
			segmentsDesc += fmt.Sprintf("%3d ", segment)
		}
		segmentsDesc = strings.TrimRight(segmentsDesc, " ")
	}

	fmt.Printf(
		"limiter status: windowTotal=%v segments=[%s]\n",
		stats.WindowTotal,
		segmentsDesc,
	)
}
