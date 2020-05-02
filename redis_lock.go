package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var (
	numWorkers int32 = 10
)

func incr() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	var lockKey = "counte_lock"
	var counterKey = "counter"

	//step 1:try to get the lock,sleep a while and contine to try lock if lock failed
	for {
		resp := client.SetNX(lockKey, 1, time.Second*5)
		lockSucc, err := resp.Result()
		if err != nil || !lockSucc {
			fmt.Println(err, "Lock result:", lockSucc)
			time.Sleep(time.Millisecond * time.Duration(rand.Int31n(numWorkers+1)))
			continue
		}
		break
	}

	//step 2:do something
	getResp := client.Get(counterKey)
	cntValue, err := getResp.Int64()
	if err == nil {
		cntValue++
		resp := client.Set(counterKey, cntValue, 0)
		_, err := resp.Result()
		if err != nil {
			fmt.Println("set value err ", err)
		}
	}
	fmt.Println("current counter is ", cntValue)

	//step 3: unlock
	delResp := client.Del(lockKey)
	unlockSucc, err := delResp.Result()
	if err == nil && unlockSucc > 0 {
		fmt.Println("unlock success")

	} else {
		fmt.Println("unlock failed", err)
	}
}

func main() {
	var wg sync.WaitGroup
	for i := int32(0); i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			incr()
		}()
	}
	wg.Wait()
}
