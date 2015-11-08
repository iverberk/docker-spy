package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"fmt"

	dockerApi "github.com/fsouza/go-dockerclient"
)

var dnsBind = flag.String("dns-bind", getopt("DNS_BIND", "0.0.0.0"), "Bind address for the DNS server")
var dnsPort = flag.String("dns-port", getopt("DNS_PORT", "53"), "Port for the DNS server")
var dnsRecursor = flag.String("dns-recursor", getopt("DNS_RECURSOR", ""), "DNS recursor for non-local addresses")
var dnsDomain = flag.String("dns-domain", getopt("DNS_DOMAIN", "localdomain"), "The domain that Docker-spy should consider local")
var dockerHost = flag.String("docker-host", getopt("DOCKER_HOST", "unix:///var/run/docker.sock"), "Address for the Docker daemon")
var dockerCertPath = flag.String("docker-cert-path", getopt("DOCKER_CERT_PATH", ""), "Location of certificates for TLS")

func getopt(name, def string) string {
	if env := os.Getenv(name); env != "" {
		return env
	}
	return def
}

func main() {

	flag.Parse()

	log.Println("Starting DNS server...")

	port, err := strconv.Atoi(*dnsPort)
	if err != nil {
		log.Fatalf("Could not convert %s to numeric type", *dnsPort)
	}

	server := &DNS{
		bind:      *dnsBind,
		port:      port,
		recursors: []string{*dnsRecursor + ":53"},
		domain:    *dnsDomain + ".",
	}

	server.Run()

	log.Println("Listening for container events...")

	var docker *dockerApi.Client
	
	if *dockerCertPath == "" {
		docker, err = dockerApi.NewClient(*dockerHost)		
	} else {
		ca := fmt.Sprintf("%s/ca.pem", *dockerCertPath)
    	cert := fmt.Sprintf("%s/cert.pem", *dockerCertPath)
    	key := fmt.Sprintf("%s/key.pem", *dockerCertPath)
    	docker, err = dockerApi.NewTLSClient(*dockerHost, cert, key, ca)
	}

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
