package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	address         = flag.String("address", "127.0.0.2:8080", "http host port to be evaluated")
	writeIterations = flag.Int("write-iterations", 1000, "number of write iterations")
	readIterations  = flag.Int("read-iterations", 1000, "number of read iterations")
	concurrency     = flag.Int("concurrency", 1, "number of goroutines to be doing writes in parallel")
)

// concurrency less than 32
var httpClient = &http.Client{
	Transport: &http.Transport{
		IdleConnTimeout:     time.Second * 60,
		MaxIdleConns:        32,
		MaxConnsPerHost:     32,
		MaxIdleConnsPerHost: 32,
	},
}

func testqps(name string, iterations int, fn func() string) (qps float64, keys []string) {
	var max time.Duration
	var min = time.Hour

	start := time.Now()
	for i := 0; i < iterations; i++ {
		iterationStart := time.Now()
		keys = append(keys, fn())
		iterationTime := time.Since(iterationStart)
		if iterationTime > max {
			max = iterationTime
		}
		if iterationTime < min {
			min = iterationTime
		}
	}

	average := time.Since(start) / time.Duration(iterations)
	qps = float64(iterations) / (float64(time.Since(start)) / float64(time.Second))
	fmt.Printf("Func %s took %s avg, %.1f QPS, %s max, %s min\n", name, average, qps, max, min)
	return qps, keys
}

func randomWrite() string {
	key := fmt.Sprintf("key-%d", rand.Intn(1000000))
	value := fmt.Sprintf("value-%d", rand.Intn(1000000))

	values := url.Values{}
	values.Set("key", key)
	values.Set("value", value)

	resp, err := httpClient.Get("http://" + (*address) + "/put?" + values.Encode())
	if err != nil {
		log.Fatalf("errror with randomWrite(): %v", err)
	}

	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return key
}

func randomRead(keys []string) string {
	key := keys[rand.Intn(len(keys))]

	values := url.Values{}
	values.Set("key", key)

	resp, err := httpClient.Get("http://" + (*address) + "/get?" + values.Encode())
	if err != nil {
		log.Fatalf("errror with randomRead(): %v", err)
	}

	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return key
}

func writePerf() []string {
	fmt.Printf("Running with %d write iterations and concurency of %d\n", *writeIterations, *concurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalQPS float64
	var allKeys []string

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			qps, keys := testqps("write", *writeIterations, randomWrite)
			mu.Lock()
			totalQPS += qps
			allKeys = append(allKeys, keys...)
			mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Printf("Total write QPS: %.1f and Number of Set Keys: %d \n", totalQPS, len(allKeys))
	return allKeys
}

func readPerf(keys []string) {
	fmt.Printf("Running with %d read iterations and concurency of %d\n", *readIterations, *concurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalQPS float64

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			qps, _ := testqps("read", *readIterations, func() string {
				return randomRead(keys)
			})
			mu.Lock()
			totalQPS += qps
			mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Printf("Total Read QPS: %.1f \n", totalQPS)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	keys := writePerf()
	readPerf(keys)

	// If we want to do concurrent writes and reads
	// go writePerf()
	// readPerf(keys)
}
