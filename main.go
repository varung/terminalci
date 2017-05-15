package main

import (
	"log"
	"runtime"
	"time"
)

func read(i int, out chan int) {
	<-time.After(75 * time.Millisecond)
	//log.Println("disk_read ", i)
	out <- i
}

func process(in chan int, out chan int) {
	for {
		x := <-in
		//log.Println("process: ", x)
		<-time.After(15 * time.Millisecond)
		if x%3 == 2 {
			go read(x, out)
		} else {
			out <- x
		}
	}
}

func make_requests(in chan int) {
	for i := 1; i < 10000; i++ {
		in <- i
	}
	close(in)
}

func main() {
	runtime.GOMAXPROCS(2)
	queue := make(chan int, 2)
	results := make(chan int, 2)
	log.SetFlags(log.Lmicroseconds)
	// check time
	log.Println("start")
	t := time.Now()
	go process(queue, results)
	go make_requests(queue)

	log.Println("reading")
	count := 0
	for {
		x := <-results
		if x == 0 {
			break
		}
		log.Println("received ", count, x)
		count++
		if time.Since(t) > time.Second {
			break
		}
	}
}
