#!/bin/bash
go get github.com/fsouza/go-dockerclient
go get  github.com/miekg/dns
go build -o docker-spy *.go
strip docker-spy
