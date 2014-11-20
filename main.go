package main

import (
	dockerApi "github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"os/signal"
)

func main() {

	log.Println("Starting DNS server...")

	server := &DNS{
		host:   "0.0.0.0",
		port:   6500,
		domain: ".",
	}

	server.Run()

	log.Println("Listening for container events...")

	docker, err := dockerApi.NewClient("unix://var/run/docker.sock")

	if err != nil {
		log.Fatal(err)
	}

	spy := &Spy{
		docker: docker,
		dns:    server,
	}

	spy.Watch()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

forever:
	for {
		select {
		case <-sig:
			log.Println("signal received, stopping")
			break forever
		}
	}
}
