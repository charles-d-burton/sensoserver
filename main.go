package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sensoserver/requests"
	"sensoserver/workers"
	"strings"

	"github.com/urfave/cli"
	"rsc.io/letsencrypt"
)

var (
	nWorkers = runtime.NumCPU()
	port     string
	database string
	queue    string
	nsqPort  string
	nsqHost  string
)

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//Start the work dispatcher
	key := os.Getenv("APIKEY")
	app := processCLI()
	app.Action = func(c *cli.Context) error {
		if c.NArg() > 0 {
			//TODO: Do something with args
		}
		log.Println("Database initialized at: ", database)
		err := workers.StartBolt(database, 0600)
		if err != nil {
			panic(err)
		}
		if c.String("ssl") == "on" {
			var m letsencrypt.Manager
			if err := m.CacheFile("letsencrypt.cache"); err != nil {
				log.Fatal(err)
			}
			log.Fatal(m.Serve())
		} else if c.String("ssl") == "off" {
			log.Println("Listening on port ", port)
			log.Fatal(http.ListenAndServe("localhost:"+port, nil))
		} else if c.String("queue") != "firebase" {
			setupQueue(queue, nsqHost, nsqPort)
		}
		return nil
	}
	workers.StartDispatcher(nWorkers, strings.Trim(key, " "))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS!\n")
	})

	http.HandleFunc("/reading", requests.Reading)
	http.HandleFunc("/register", requests.Register)
	http.HandleFunc("/register/jointopic", requests.JoinTopic)
	http.HandleFunc("/register/leavetopic", requests.LeaveTopic)
	http.HandleFunc("/register/refreshtoken", requests.RefreshToken)
	http.HandleFunc("/register/registerdevice", requests.RegisterDevice)

	app.Run(os.Args)

	//conn.Close()
}

func setupQueue(queueType, host, port string) {
	if queueType == "both" || queueType == "nsq" {
		workers.SetQueueType(queueType)
		workers.SetupNSQ(host, port)
	}
}

func processCLI() *cli.App {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ssl, s",
			Value: "on",
			Usage: "set to 'off' to disable ssl",
		},
		cli.StringFlag{
			Name:        "port, p",
			Value:       "8901",
			Usage:       "Set listening port for unencrypted traffic\n                 Only used if --ssl is set to off",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "database, d",
			Value:       "./senso.db",
			Usage:       "Directory and file to save the persistence database",
			Destination: &database,
		},
		cli.StringFlag{
			Name:        "queue, q",
			Value:       "firebase",
			Usage:       "Select either \"firebase\", \"nsq\", or \"both\".  Firebase sends to firebase cloud messaging, NSQ uses nsq.io",
			Destination: &queue,
		},
		cli.StringFlag{
			Name:        "nsqport, np",
			Value:       "4150",
			Usage:       "Set the port to connect to nsq",
			Destination: &nsqPort,
		},
		cli.StringFlag{
			Name:        "nsqhost, nh",
			Value:       "localhost",
			Usage:       "Set the NSQ host to connect to",
			Destination: &nsqHost,
		},
	}
	return app
}

// GetLocalIP returns the non loopback local IP of the host
func getLocalIPS() []net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	var ips []net.IP
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.String() != "127.0.1.1" {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips
}
