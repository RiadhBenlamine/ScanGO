package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)


type Result struct {
	Port string
	Open bool
	Proto string
}

type Config struct {
	Target  string
	Ports   string
	Threads int
	Udp     bool
}
func banner() {
	fmt.Println("=====================================================================")
	fmt.Println("                            ScanGO                      ")
	fmt.Println("=====================================================================")

}

func handlePorts(ports string) []string {
	if ports == "-" {
		return genRange(1, 65535)
	}
	if strings.Contains(ports, ",") {
		return strings.Split(ports, ",")
	}
	if strings.Contains(ports, "-") {
		parts := strings.Split(ports, "-")
		if len(parts) != 2 {
			return nil
		}
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil ||
			start < 1 || end > 65535 || start > end {
			fmt.Println("Invalid port range")
			return nil
		}
		return genRange(start, end)
	}
	return []string{ports}
}

func genRange(start, end int) []string {
	list := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		list = append(list, strconv.Itoa(i))
	}
	return list
}

func scan(addr string, result chan<- Result) {

	conn, err := net.DialTimeout("tcp", addr, time.Millisecond*950)
	getPort := strings.Split(addr, ":")[1]
	if err != nil {
		result <- Result{Port: getPort, Open: false}
		return
	}
	err = conn.Close()
	if err != nil {
		log.Println(err)
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
	banner()
	config := &Config{}
	cmd := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "target",
				Required:    true,
				Usage:       "target host",
				Aliases:     []string{"t"},
				Destination: &config.Target,
			},
			&cli.StringFlag{
				Name:        "ports",
				Usage:       "target ports",
				Aliases:     []string{"p"},
				Destination: &config.Ports,
				Value:       "445,5432,80,135",
			},
			&cli.IntFlag{
				Name:        "threads",
				Usage:       "number of threads, Default: 1",
				Aliases:     []string{"T"},
				Destination: &config.Threads,
				Value:       1,
			},
			&cli.BoolFlag{
				Name:        "UDP",
				Aliases:     []string{"U"},
				Value:       false,
				Destination: &config.Udp,
			},
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	jobs := make(chan string)
	results := make(chan Result)
	var wg sync.WaitGroup
	host := config.Target
	if !config.Udp {
		fmt.Println("Preforming TCP scan.")
	}else{
		fmt.Println("Preforming UDP scan.")
	}
	if host == "" {
		log.Fatal("No target host specified")
	}
	fmt.Println("Target host:", host)

	list := handlePorts(config.Ports)
	fmt.Println("Scanning ", len(list), "ports...")

	for range config.Threads {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}
	start := time.Now()
	go func() {
		for _, port := range list {
			jobs <- net.JoinHostPort(host, port)
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()
	var counter int64
	for r := range results {
		if r.Open {
			counter++
			fmt.Println(r.Port + " open")

		}
	}
	end := time.Now()
	total := end.Sub(start)
	fmt.Println("Time taken:", total)
	fmt.Println("Total ports open:", counter)
	fmt.Println("Ports closed/filtered:", len(list))

}
