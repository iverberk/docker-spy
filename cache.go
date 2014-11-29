package main

import (
	"regexp"
)

type Record struct {
	ip   string
	arpa string
	fqdn string
}

type Cache struct {
	records map[string]*Record
}

func (c Cache) Set(id string, r *Record) {
	c.records[id] = r
}

// Provides lookups based on fqdn or ip
func (c Cache) Get(id string) (*Record, bool) {

	var reverseLookup = regexp.MustCompile("^.*\\.in-addr\\.arpa\\.$")

	if reverseLookup.MatchString(id) {
		for _, record := range c.records {
			if record.arpa == id {
				return record, true
			}
		}
	}

	if record, ok := c.records[id]; ok {
		return record, true
	}

	return nil, false
}

func (c Cache) Remove(id string) {
	delete(c.records, id)
}
