package main

import (
	"log"
)

type Record struct {
	ip string
}

type Cache struct {
	records map[string]*Record
}

func (c Cache) Set(hostname string, r *Record) {
	c.records[hostname] = r

	log.Printf("records: %v", c.records)
}

func (c Cache) Get(hostname string) (*Record, bool) {
	if record, ok := c.records[hostname]; ok {
		return record, true
	}

	return nil, false
}

func (c Cache) Remove(hostname string) {
	delete(c.records, hostname)
}
