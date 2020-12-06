// limiter introduces a pattern to use channels to control the level of concurrency of requests passing through an api.
// This pattern can help to reduce the risk of bombarding dependent api's with more requests than they can handle.
// Incoming requests will be funneled into a queue and a worker pool will process the requests with the number of workers
// being the number of concurrent requests that are allowed.  This example also includes handler functions for stopping,
// starting, and modifying the number of workers in the worker pool.
//
// *** IMPORTANT ***
//
// The application using this pattern needs to be carefully architected to ensure proper error and panic recovery so that
// queued requests are not lost in the case of a panic.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"sync"

	"github.com/gorilla/mux"

	"examples/patterns"
)

//  channels and waitgroup must be included in the controller struct to be able to stop, start, and update
type controller struct {
	queue chan int                // job queue
	done  chan struct{}           // channel to signal workers to stop processing requests
	cl    *patterns.ClientWrapper // http.client
	limit *sync.WaitGroup         // anytime a waitgroup is added to a controller struct it needs to be a pointer
}

func main() {
	// create http.client
	tr := patterns.NewTransportWrapper()
	cl := patterns.NewClientWrapper(patterns.Transport(tr))

	// create channels
	work := make(chan int, 10)
	done := make(chan struct{})

	var limit sync.WaitGroup

	// initialize controller
	ctrl := controller{
		queue: work,
		done:  done,
		cl:    cl,
		limit: &limit,
	}

	// http server
	go ctrl.run()

	// producer, this sends a finite number of jobs to the channel
	// the real implementation would send incoming requests to the channel

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			// send job to channel / queue
			work <- i
		}
	}()

	ctrl.wgroup() // starts the worker group with the default number of workers

	wg.Wait()

	close(work) // all work is done
}

func (c *controller) run() {
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/stop", c.stop())
	r.Handle("/start", c.start())
	r.Handle("/worker/add", c.addWorker())

	log.Fatal(http.ListenAndServe(":4000", r))

}

// wgroup starts the worker pool with the default number of workers
func (c *controller) wgroup() {
	for i := 0; i < 5; i++ {
		c.limit.Add(1)
		go c.startWorker()
	}
}

// startWorker adds a single worker to the worker pool
func (c *controller) startWorker() {
	defer c.limit.Done()

	for ww := range c.queue {
		select {
		case <-c.done:
			fmt.Println("send on done")
			c.request(ww)
			return
		default:

			c.request(ww)

		}

	}
}

// stop sends a signal to all workers in the pool to complete tasks in flight and terminate, stopping consumption from the work queue.
func (c *controller) stop() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.done <- struct{}{}
		close(c.done)

		fmt.Println("sent signal to done chan")
	}
}

// start restarts consumption from the work queue by reinitializing the done channel and restarting the worker pool.
func (c *controller) start() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.done = make(chan struct{})

		go c.wgroup()
		fmt.Println("restarted consumer")

	}
}

// addWorker will add a single worker to the worker pool
func (c *controller) addWorker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.limit.Add(1)
		go c.startWorker()
	}
}

// request is the work function
func (c *controller) request(i int) {
	resp, err := c.cl.Cl.Get("http://localhost:3000/health")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("request: %d %v", i, string(b))
}
