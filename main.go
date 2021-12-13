package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var domain string
var ttl int64
var max int

func main() {
	flag.StringVar(&domain, "d", "", "ns domain")
	flag.Int64Var(&ttl, "t", 30, "name ttl (second)")
	flag.IntVar(&max, "m", 20, "max records per name")
	flag.Parse()
	if domain == "" {
		flag.Usage()
		os.Exit(0)
	}
	println(fmt.Sprintf(">>>DNSLOG<<< domain:%s ttl:%d max:%d", domain, ttl, max))
	db := newDB(30 * time.Second, 20)
	handler := &dnsHandler{
		domain: domain,
		db:     db,
	}
	stop := make(chan bool)
	go startDnsServer(handler, stop)
	go startHttpServer(db, domain, stop)
	<- stop
}
