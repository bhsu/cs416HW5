/*
 [worker-incoming ip:port] : the IP:port address that workers use to connect to the server
 [client-incoming ip:port] : the IP:port address that clients use to connect to the server
 go run worker.go [server ip:port]
 go run server.go [worker-incoming ip:port] [client-incoming ip:port]
*/

package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

var (
	worker_incoming_ip_port string
	client_incoming_ip_port string
	workerIpArray           []string
	client                  *rpc.Client
)

// Resource server type (for RPC) with client
type MServer int

// Resource server type (for RPC) with worker
type Worker int

type MWebsiteReq struct {
	URI              string // URI of the website to measure
	SamplesPerWorker int    // Number of samples, >= 1
}

// Request that client sends in RPC call to MServer.GetWorkers
type MWorkersReq struct {
	SamplesPerWorker int // Number of samples, >= 1
}

// Response to:
// MServer.MeasureWebsite:
//   - latency stats per worker to a *URI*
//   - (optional) Diff map
// MServer.GetWorkers
//   - latency stats per worker to the *server*
type MRes struct {
	Stats map[string]LatencyStats    // map: workerIP -> LatencyStats
	Diff  map[string]map[string]bool // map: [workerIP x workerIP] -> True/False
}

type WorkerRes struct {
	WorkerIp string
	Min      int
	Median   int
	Max      int
	Md5Value string
}

// A stats struct that summarizes a set of latency measurements to an
// internet host.
type LatencyStats struct {
	Min    int // min measured latency in milliseconds to host
	Median int // median measured latency in milliseconds to host
	Max    int // max measured latency in milliseconds to host
}

func main() {

	done := make(chan int)
	args := os.Args[1:]

	// Missing command line args.
	if len(args) != 2 {
		fmt.Println("Usage: go run server.go [worker-incoming ip:port] [client-incoming ip:port]")
		return
	}

	worker_incoming_ip_port = args[0] // setting command line args
	client_incoming_ip_port = args[1]
	fmt.Println("\nMain funtion: commandline args check worker_incoming_ip_port: ", worker_incoming_ip_port, "client_incoming_ip_port: ", client_incoming_ip_port)

	go InitServerWorkerRPC() // start listening for worker

	go InitServerClient() // start listening for client

	<-done
}

// start listening for client request
func InitServerClient() {
	fmt.Println("\nfunc InitServerClient: start listening for client")
	cServer := rpc.NewServer()
	c := new(MServer)
	cServer.Register(c)

	l, err := net.Listen("tcp", client_incoming_ip_port)
	ErrorCheck("\nfunc InitServerClient:", err, false)
	for {
		conn, err := l.Accept()
		ErrorCheck("\nfunc InitServerClient:", err, false)
		go cServer.ServeConn(conn)
	}
}

// FIXME get rid of this func, it is here for cserver to complie for now, this will need to be replaced later on for A5 specs
// rpc from client requesting web stats
func (c *MServer) MeasureWebsite(m MWebsiteReq, reply *MRes) error {
	return nil
}

// Create RPC client for contacting the worker.
func getRPCClientServer(ip string) *rpc.Client {

	raddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		ErrorCheck("\nfunc getRPCClientServer:", err, false)
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		ErrorCheck("\nfunc getRPCClientServer:", err, false)
	}
	client := rpc.NewClient(conn)
	return client
}

// start listening for worker request
func InitServerWorkerRPC() {
	fmt.Println("\nfunc InitServerWorkerRPC: start listening for worker...")
	wServer := rpc.NewServer()
	w := new(Worker)
	wServer.Register(w)

	l, err := net.Listen("tcp", worker_incoming_ip_port)
	ErrorCheck("\nfucn InitServerWorkerRPC:", err, false)
	for {
		conn, err := l.Accept()
		ErrorCheck("\nfucn InitServerWorkerRPC:", err, false)
		go wServer.ServeConn(conn)
	}
}

// receive worker ip via RPC
func (w *Worker) ReceiveWorkerIp(ip string, reply *bool) error {
	fmt.Println("\nfunc ReceiveWorkerIp: Received:", ip, "\n")
	workerIpArray = append(workerIpArray, ip) // add ip to the worker ip array
	*reply = true
	return nil
}

func ErrorCheck(msg string, err error, exit bool) {
	if err != nil {
		log.Println(msg, err)
		if exit {
			os.Exit(-1)
		}
	}
}
