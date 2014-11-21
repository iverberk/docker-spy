package main

import (
	"log"
	"os"
	"os/signal"

	dockerApi "github.com/fsouza/go-dockerclient"
)

func main() {

	log.Println("Starting DNS server...")

	server := &DNS{
		host:      "0.0.0.0",
		port:      6500,
		recursors: []string{"130.115.1.1:53", "130.115.15.2:53"},
		domain:    "localdomain.",
	}

	server.Run()

	log.Println("Listening for container events...")

	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		host = "unix:///var/run/docker.sock"
	}

	docker, err := dockerApi.NewClient(host)

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
