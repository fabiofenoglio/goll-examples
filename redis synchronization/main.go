package main

import (
	"fmt"
	"math/rand"
	"time"

	goll "github.com/fabiofenoglio/goll"
	gollredis "github.com/fabiofenoglio/goll-redis"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

func main() {
	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		// set your redis connection parameters here.
		Addr: "localhost:6379",
	})
	defer client.Close()

	pool := goredis.NewPool(client)

	adapter, err := gollredis.NewRedisSyncAdapter(&gollredis.Config{
		Pool:      pool,
		MutexName: "redisAdapterTest",
	})

	if err != nil {
		panic(err)
	}

	// create an instance of LoadLimiter
	// accepting a max of 1000 over 20 seconds
	newLimiter, _ := goll.New(&goll.Config{
		MaxLoad:     1000,
		WindowSize:  20 * time.Second,
		SyncAdapter: adapter,
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
		submitResult, err := limiter.Submit(requestedLoad)
		if err != nil {
			panic(fmt.Errorf("error submitting: %w", err))
		}

		if submitResult.Accepted {
			accepted += requestedLoad
			fmt.Printf("request for load of %v was accepted\n", requestedLoad)

		} else if submitResult.RetryInAvailable {
			fmt.Printf("request for load of %v was rejected, asked to wait for %v ms\n", requestedLoad, submitResult.RetryIn.Milliseconds())

			// wait, then resubmit
			time.Sleep(submitResult.RetryIn)

			requested += requestedLoad
			submitResult, err = limiter.Submit(requestedLoad)

			if err != nil {
				panic(fmt.Errorf("error resubmitting: %w", err))
			}
			if submitResult.Accepted {
				fmt.Printf("resubmitted request for load of %v was accepted\n", requestedLoad)
				accepted += requestedLoad
			} else {
				panic("waited the required amount of time but the request was rejected again :(")
			}

		} else {
			fmt.Printf("request for load of %v was rejected with no indications on the required delay before resubmitting\n", requestedLoad)
		}

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
