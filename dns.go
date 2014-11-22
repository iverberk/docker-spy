package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"strconv"
)

type DNS struct {
	bind      string
	port      int
	domain    string
	recursors []string
	cache     Cache
}

func (s *DNS) Run() {

	s.cache.records = make(map[string]*Record)

	mux := dns.NewServeMux()

	srvUDP := &dns.Server{
		Addr:    s.bind + ":" + strconv.Itoa(s.port),
		Net:     "udp",
		Handler: mux,
	}

	srvTCP := &dns.Server{
		Addr:    s.bind + ":" + strconv.Itoa(s.port),
		Net:     "tcp",
		Handler: mux,
	}

	mux.HandleFunc(".", s.handleDNSExternal)
	mux.HandleFunc(s.domain, s.handleDNSInternal)

	go func() {
		err := srvUDP.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		err := srvTCP.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func (s *DNS) handleDNSInternal(w dns.ResponseWriter, req *dns.Msg) {

	q := req.Question[0]

	if q.Qtype == dns.TypeA && q.Qclass == dns.ClassINET {
		if record, ok := s.cache.Get(q.Name); ok {

			log.Printf("Found internal record for %s", q.Name)

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

			log.Printf("sent: %v", m)
			return
		}

		log.Printf("No internal record found for %s", q.Name)
		dns.HandleFailed(w, req)
	}

	log.Printf("Only handling type A requests, skipping")
	dns.HandleFailed(w, req)
}

func (s *DNS) handleDNSExternal(w dns.ResponseWriter, req *dns.Msg) {

	network := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		network = "tcp"
	}

	c := &dns.Client{Net: network}
	var r *dns.Msg
	var err error
	for _, recursor := range s.recursors {
		r, _, err = c.Exchange(req, recursor)
		if err == nil {
			if err := w.WriteMsg(r); err != nil {
				log.Printf("DNS lookup failed %v", err)
			}
			log.Printf("Found external record for " + req.Question[0].Name)
			return
		}
	}

	dns.HandleFailed(w, req)
}
