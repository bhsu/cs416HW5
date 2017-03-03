package main

/*
Azure listens on private ip, and caller will need to call on public ip
*/

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

var (
	server_ip_port string
	workerport_RPC string
	public_ip      string
	private_ip     string
	client         *rpc.Client // RPC client.
)

// Resource server type.
type Worker int

type MWebsiteReq struct {
	URI              string // URI of the website to measure
	SamplesPerWorker int    // Number of samples, >= 1
}

// Response to:
// MServer.MeasureWebsite:
//   - latency stats per worker to a *URI*
//   - (optional) Diff map
// MServer.GetWorkers
//   - latency stats per worker to the *server*
type MRes struct {
	Stats map[string]LatencyStats // map: workerIP -> LatencyStats
	//Diff  map[string]map[string]bool // map: [workerIP x workerIP] -> True/False
}

// A stats struct that summarizes a set of latency measurements to an
// internet host.
type LatencyStats struct {
	Min    int // min measured latency in milliseconds to host
	Median int // median measured latency in milliseconds to host
	Max    int // max measured latency in milliseconds to host
}

type WorkerRes struct {
	WorkerIp string
	Min      int
	Median   int
	Max      int
	Md5Value string
}

func main() {

	args := os.Args[1:]
	done := make(chan int)

	// Missing command line args.
	if len(args) != 1 {
		fmt.Println("Usage: go run server.go [worker-incoming ip:port] [client-incoming ip:port]")
		return
	}

	server_ip_port = args[0]
	workerport_RPC = splitIpToGetPort(server_ip_port)
	fmt.Println("\nMain funtion: commandline args check server_ip_port: ", server_ip_port)

	public_ip = getPublicIp()     // gets external ip and send it to server to store
	client = getRPCClientWorker() // Create RPC client for contacting the server.
	sendWorkerIpToServer(client)  // send local ip to server to store

	<-done
}

func InitWorkerServerRPC() {
	fmt.Println("\nfunc InitWorkerServerRPC: start listening for server rpc")
	wServer := rpc.NewServer()
	w := new(Worker)
	wServer.Register(w)

	ip := getPrivateIp() + ":" + workerport_RPC // listens on private ip
	fmt.Println("func InitWorkerServerRPC: private ip", ip)

	l, err := net.Listen("tcp", ip)
	checkError("\nfunc InitWorkerServerRPC:", err, false)
	for {
		fmt.Println("\nfunc InitWorkerServerRPC: start listening for server rpc...")
		conn, err := l.Accept()
		checkError("\nfunc InitWorkerServerRPC:", err, false)
		go wServer.ServeConn(conn)
	}
}

func getRPCClientWorker() *rpc.Client {

	raddr, err := net.ResolveTCPAddr("tcp", server_ip_port)
	fmt.Println("\nfunc getRPCClientWorker: workerport_RPC:", workerport_RPC)
	fmt.Println("\nfunc getRPCClientWorker server_ip_port:", server_ip_port)
	if err != nil {
		checkError("\nfunc getRPCClientWorker:", err, false)
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		checkError("\nfunc getRPCClientWorker:", err, false)
	}
	client := rpc.NewClient(conn)
	return client
}

// sends public ip to server
func sendWorkerIpToServer(client *rpc.Client) (bool, error) {
	fmt.Println("\nfunc sendWorkerIptoServer: sending ip to server to store..")
	var reply bool
	var err error
	err = client.Call("Worker.ReceiveWorkerIp", public_ip, &reply) // send public to the server to store
	checkError("\nfunc sendWorkerIptoServer:", err, false)
	client.Close()           // close the client, so we can start listening
	go InitWorkerServerRPC() // start listening
	return reply, err
}

// Get preferred outbound ip of this machine, this ip will be used as listening port
func getPrivateIp() string {
	fmt.Println("\nfunc getPrivateIp: getting private ip")
	var ip string
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	ip = localAddr[0:idx]
	fmt.Println("func prviateIp:", ip, "\n")
	return ip
}

// gets external ip and send it to server to store
func getPublicIp() string {
	fmt.Println("\nfunc getPublicIp: getting public ip")
	var ip string
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	defer resp.Body.Close()
	//io.Copy(os.Stdout, resp.Body)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newStr := buf.String()
	ip = strings.TrimSpace(newStr)
	fmt.Println("getPublicIp", ip)
	return ip
}

// split the ip port
func splitIpToGetPort(ip string) string {
	s := strings.Split(ip, ":")
	return s[1]
}

// error checking
func checkError(msg string, err error, exit bool) {
	if err != nil {
		log.Println(msg, err)
		if exit {
			os.Exit(-1)
		}
	}
}

// FIXME get rid of this func, it is here for cserver to complie for now, this will need to be replaced later on for A5 specs
func (w *Worker) GetWeb(m MWebsiteReq, reply *WorkerRes) error {

	return nil
}
