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
	bind := s.bind + ":" + strconv.Itoa(s.port)

	srvUDP := &dns.Server{
		Addr:    bind,
		Net:     "udp",
		Handler: mux,
	}

	srvTCP := &dns.Server{
		Addr:    bind,
		Net:     "tcp",
		Handler: mux,
	}

	mux.HandleFunc(s.domain, s.handleDNSInternal)
	mux.HandleFunc("in-addr.arpa.", s.handleReverseDNSLookup)
	mux.HandleFunc(".", s.handleDNSExternal)

	go func() {
		log.Printf("Binding UDP listener to %s", bind)

		err := srvUDP.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Printf("Binding TCP listener to %s", bind)

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

		if recursor == "" {
			log.Printf("Found empty recursor")
			continue
		}

		log.Printf("Forwarding request to external recursor for: %s", req.Question[0].Name)

		r, _, err = c.Exchange(req, recursor)
		if err == nil {
			if err := w.WriteMsg(r); err != nil {
				log.Printf("DNS lookup failed %v", err)
			}
			return
		}
	}

	dns.HandleFailed(w, req)
}

func (s *DNS) handleReverseDNSLookup(w dns.ResponseWriter, req *dns.Msg) {

	q := req.Question[0]

	if q.Qtype == dns.TypePTR && q.Qclass == dns.ClassINET {

		if record, ok := s.cache.Get(q.Name); ok {

			log.Printf("Found internal record for %s", q.Name)

			m := new(dns.Msg)
			m.SetReply(req)
			rr_header := dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    0,
			}

			a := &dns.PTR{rr_header, record.fqdn}
			m.Answer = append(m.Answer, a)
			w.WriteMsg(m)

			return

		}

		log.Printf("Forwarding request to external recursor for: %s", q.Name)

		// Forward the request
		s.handleDNSExternal(w, req)
	}

	dns.HandleFailed(w, req)
}
