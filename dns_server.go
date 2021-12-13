package main

import (
	"github.com/miekg/dns"
	"net"
	"strings"
)

var localIP = net.ParseIP("127.0.0.1").To4()

func startDnsServer(handler dns.Handler, stop chan bool) {
	s := &dns.Server{
		Addr: ":53",
		Net: "udp",
		Handler: handler,
	}
	if err := s.ListenAndServe(); err != nil {
		println("ERROR [DNS] -" + err.Error())
		stop <- true
	}
}

func getIP(addr net.Addr) string {
	udpAddr := addr.(*net.UDPAddr)
	return udpAddr.IP.String()
}

type dnsHandler struct {
	domain string
	db *db
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	resp := dns.Msg{}
	resp.SetReply(r)
	for _, q := range resp.Question {
		if q.Qtype !=  dns.TypeA {
			continue
		}
		a := h.handle(getIP(w.RemoteAddr()), q)
		if a == nil {
			continue
		}
		resp.Answer = append(resp.Answer, a)
	}
	w.WriteMsg(&resp)
}

func (h *dnsHandler) handle(ip string, q dns.Question) dns.RR {
	qname := strings.TrimSuffix(q.Name, ".")
	name := strings.TrimSuffix(qname, h.domain)
	if name == "" || name == "." {
		return nil
	}
	name = strings.TrimSuffix(name, ".")
	dotIndex := strings.LastIndex(name, ".")
	if dotIndex != -1 {
		name = name[dotIndex+1:]
	}
	h.db.AddRecord(name, qname, ip)
	return &dns.A{
		Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 86400},
		A:   localIP,
	}
}
