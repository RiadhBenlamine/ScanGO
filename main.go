package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Result struct {
	Port string
	Open bool
}

func scan(addr string, result chan<- Result) {
	conn, err := net.Dial("tcp", addr)
	getPort := strings.Split(addr, ":")[1]
	defer conn.Close()
	if err != nil {
		result <- Result{Port: getPort, Open: false}
	}
	result <- Result{Port: getPort, Open: true}
}
func worker(jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for addr := range jobs {
		scan(addr, results)
	}
}

func main() {
	jobs := make(chan string)
	results := make(chan Result)
	var wg sync.WaitGroup
	host := "scname.nmap.org"
	list := []int{
		80,
		8080,
		443,
		5000,
		3000,
		5321,
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}
	go func() {
		for _, port := range list {
			jobs <- net.JoinHostPort(host, strconv.Itoa(port))
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Println(r)
	}

}
