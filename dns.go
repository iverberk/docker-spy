package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"strconv"
	"strings"
)

type DNS struct {
	host   string
	port   int
	domain string
	cache  Cache
}

func (s *DNS) Run() {

	s.cache.records = make(map[string]*Record)

	mux := dns.NewServeMux()

	srv := &dns.Server{
		Addr:    s.host + ":" + strconv.Itoa(s.port),
		Net:     "udp",
		Handler: mux,
	}

	mux.HandleFunc(".", s.handleDNSRequest)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *DNS) handleDNSRequest(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]

	name := strings.TrimSuffix(q.Name, s.domain)

	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
		if record, ok := s.cache.Get(name); ok {

			log.Printf("Found record for %s", name)

			m := new(dns.Msg)
			m.SetReply(req)
			rr_header := dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    0,
			}
			a := &dns.A{rr_header, net.ParseIP(record.ip)}
			m.Answer = append(m.Answer, a)
			w.WriteMsg(m)
			return
		}

		log.Printf("No record found for %s", name)
		dns.HandleFailed(w, req)
	}
}
