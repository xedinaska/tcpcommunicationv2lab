package main

import (
	"flag"
	"github.com/xedinaska/tcpcommunicationv2/serverapp/tcp"
	log "github.com/sirupsen/logrus"
)

const connHost = "localhost"

func main() {
	portPtr := flag.Int("port", 3333, "use -port to provide port that server should listen (3333 by default)")
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	if err := tcp.NewServer(connHost, *portPtr).Start(); err != nil {
		log.Fatalf("[FATAL] failed to start TCP server: `%s`", err.Error())
	}

	log.Printf("[INFO]server successfully started and accepting connections on port %d", *portPtr)

	//prevent app closing using infinite loop
	for {
		select {}
	}
}
