package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gammazero/deque"
)

func main() {
    port_quanitity := flag.Int("ports", 0, "Number of distinct ports needed")
    starting_port := flag.Int("start", 9000, "Port to start port range from (default: 9000)")

    flag.Parse()

    if *port_quanitity <= 0 {
        panic("number of ports must be postitve")
    }

    if *starting_port + *port_quanitity > 65535 {
        panic("starting port + number of ports must not exceed the maximum port (65535)")
    }

    log.Printf("Creating queue with %d ports\n", *port_quanitity)
    portQueue := deque.New[int](*port_quanitity)

    pqh := PortQueueHandler{
        PortQueue: portQueue,
    }

    pqh.init(*starting_port, *port_quanitity)

    http.HandleFunc("/checkout", pqh.checkout)
    http.HandleFunc("/checkin", pqh.checkin)

    http.ListenAndServe("127.0.0.1:8080", nil)
}
