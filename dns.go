package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"strconv"
)

type DNS struct {
	host      string
	port      int
	domain    string
	recursors []string
	cache     Cache
}

func (s *DNS) Run() {

	s.cache.records = make(map[string]*Record)

	mux := dns.NewServeMux()

	srv := &dns.Server{
		Addr:    s.host + ":" + strconv.Itoa(s.port),
		Net:     "udp",
		Handler: mux,
	}

	mux.HandleFunc(".", s.handleDNSExternal)
	mux.HandleFunc(s.domain, s.handleDNSInternal)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *DNS) handleDNSInternal(w dns.ResponseWriter, req *dns.Msg) {

	q := req.Question[0]

	log.Printf("Internal DNS request received for " + q.Name)

	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
		if record, ok := s.cache.Get(q.Name); ok {

			log.Printf("Found record for %s", q.Name)

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

		log.Printf("No record found for %s", q.Name)
		dns.HandleFailed(w, req)
	}
}

func (s *DNS) handleDNSExternal(w dns.ResponseWriter, req *dns.Msg) {

	log.Printf("External DNS request received for " + req.Question[0].Name)

	c := &dns.Client{Net: "udp"}
	var r *dns.Msg
	var err error
	for _, recursor := range s.recursors {
		r, _, err = c.Exchange(req, recursor)
		if err == nil {
			if err := w.WriteMsg(r); err != nil {
				log.Printf("DNS lookup failed %v", err)
			}
			return
		}
	}

	m := &dns.Msg{}
	m.SetReply(req)
	m.RecursionAvailable = true
	m.SetRcode(req, dns.RcodeServerFailure)
	w.WriteMsg(m)
}
