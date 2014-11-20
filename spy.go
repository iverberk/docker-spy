package main

import (
	dockerApi "github.com/fsouza/go-dockerclient"
	"log"
)

type Spy struct {
	docker *dockerApi.Client
	dns    *DNS
}

func (s *Spy) Watch() {

	events := make(chan *dockerApi.APIEvents)
	s.docker.AddEventListener(events)

	go s.readEventStream(events)

}

func (s *Spy) readEventStream(events chan *dockerApi.APIEvents) {
	for msg := range events {
		switch msg.Status {
		case "start":
			container := s.inspectContainer(msg.ID)

			name := container.Config.Hostname + "." + container.Config.Domainname
			log.Printf("Adding cache record for %v", name)

			s.dns.cache.Set(name, &Record{container.NetworkSettings.IPAddress})
		case "die":
			container := s.inspectContainer(msg.ID)

			name := container.Config.Hostname + "." + container.Config.Domainname

			log.Printf("Removing cache record for %v", name)
			s.dns.cache.Remove(name)
		}
	}
}

func (s *Spy) inspectContainer(id string) *dockerApi.Container {

	container, err := s.docker.InspectContainer(id)
	if err != nil {
		log.Fatalf("Unable to inspect container: %s", id[:12])
		return nil
	}

	return container
}
