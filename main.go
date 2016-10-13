package main

import (
	"fmt"
	"log"
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
		}
		return nil
	}
	workers.StartDispatcher(nWorkers, strings.Trim(key, " "))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS!\n")
	})

	http.HandleFunc("/reading", requests.Reading)
	http.HandleFunc("/register", requests.Register)
	http.HandleFunc("/jointopic", requests.JoinTopic)
	http.HandleFunc("/leavetopic", requests.LeaveTopic)
	http.HandleFunc("/refreshtoken", requests.RefreshToken)
	http.HandleFunc("/registerdevice", requests.RegisterDevice)

	app.Run(os.Args)

	//conn.Close()
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
	}
	return app
}
