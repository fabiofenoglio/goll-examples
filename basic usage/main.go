package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	goll "github.com/fabiofenoglio/goll"
)

func main() {
	// create an instance of LoadLimiter
	// accepting a max of 1000 over 10 seconds
	newLimiter, _ := goll.New(&goll.Config{
		MaxLoad:    1000,
		WindowSize: 10 * time.Second,
	})

	// not caring for multi-tenancy right now,
	// so let's switch to a single-tenant interface
	limiter := newLimiter.ForTenant("test")

	// we'll gather some stats to see how our boy performs
	startedAt := time.Now()
	requested := uint64(0)
	accepted := uint64(0)

	for i := 0; i < 100; i++ {
		// require a random amount of load from 20 to 50
		// to simulate various kind of requests
		requestedLoad := uint64(rand.Intn(50-20) + 20)

		requested += requestedLoad

		// submit the request to the limiter
		submitResult, _ := limiter.Submit(requestedLoad)

		if submitResult.Accepted {
			accepted += requestedLoad
			fmt.Printf("request for load of %v was accepted\n", requestedLoad)

		} else if submitResult.RetryInAvailable {
			fmt.Printf("request for load of %v was rejected, asked to wait for %v ms\n", requestedLoad, submitResult.RetryIn.Milliseconds())

			// wait, then resubmit
			time.Sleep(submitResult.RetryIn)

			requested += requestedLoad
			submitResult, _ = limiter.Submit(requestedLoad)

			if submitResult.Accepted {
				fmt.Printf("resubmitted request for load of %v was accepted\n", requestedLoad)
				accepted += requestedLoad
			} else {
				panic("waited the required amount of time but the request was rejected again :(")
			}

		} else {
			fmt.Printf("request for load of %v was rejected with no indications on the required delay before resubmitting\n", requestedLoad)
		}

		// print a recap of the limiter status
		displayLimiterStatus(limiter)

		// sleep for a random amount of time from 0 to 500ms
		// to simulate random requests pattern
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}

	// check if the limiter really did limit.
	// we expect a "total accepted" around 100/sec (1000 over 10 seconds)
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
