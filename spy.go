package main

import (
	dockerApi "github.com/fsouza/go-dockerclient"
	"github.com/miekg/dns"
	"log"
	"regexp"
)

type Spy struct {
	docker *dockerApi.Client
	dns    *DNS
}

func (s *Spy) Watch() {

	s.registerRunningContainers()

	events := make(chan *dockerApi.APIEvents)
	s.docker.AddEventListener(events)

	go s.readEventStream(events)
}

func (s *Spy) registerRunningContainers() {
	containers, err := s.docker.ListContainers(dockerApi.ListContainersOptions{})
	if err != nil {
		log.Fatalf("Unable to register running containers: %v", err)
	}
	for _, listing := range containers {
		s.mutateContainerInCache(listing.ID, listing.Status)
	}
}

func (s *Spy) readEventStream(events chan *dockerApi.APIEvents) {
	for msg := range events {
		s.mutateContainerInCache(msg.ID, msg.Status)
	}
}

func (s *Spy) mutateContainerInCache(id string, status string) {

	container, err := s.docker.InspectContainer(id)
	if err != nil {
		log.Printf("Unable to inspect container %s, skipping", id)
		return
	}

	name := container.Config.Hostname + "." + container.Config.Domainname + "."

	var running = regexp.MustCompile("start|^Up.*$")
	var stopping = regexp.MustCompile("die")

	switch {
	case running.MatchString(status):
		log.Printf("Adding record for %v", name)
		arpa, err := dns.ReverseAddr(container.NetworkSettings.IPAddress)
		if err != nil {
			log.Printf("Unable to create ARPA address. Reverse DNS lookup will be unavailable for this container.")
		}
		s.dns.cache.Set(name, &Record{
			container.NetworkSettings.IPAddress,
			arpa,
			name,
		})
	case stopping.MatchString(status):
		log.Printf("Removing record for %v", name)
		s.dns.cache.Remove(name)
	}
}
