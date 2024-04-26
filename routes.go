package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/gammazero/deque"
)

type PortQueueHandler struct {
    PortQueue *deque.Deque[int]
}

type CheckoutResponse struct {
    Port int `json:"port"`
    Message string `json:"message"`
}

type CheckinResponse struct {
    Message string `json:"message"`
}

type Request struct {
    Port int `json:"port"`
}


func (qh *PortQueueHandler) checkout(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    queue := qh.PortQueue

    if queue.Len() == 0 {
	res := CheckoutResponse{
	    Message: "No ports left in queue",
	}
	resBytes, _ := json.Marshal(res)
	w.Write(resBytes) 
	log.Println("no more ports in queue")	
	return
    }

    port := queue.PopFront()
    res := CheckoutResponse{
	Port: port,
	Message: "Successfully checked out port",
    }
    resBytes, _ := json.Marshal(res)

    w.Write(resBytes) 
    log.Printf("returned port %d\n", port)
}

func (qh *PortQueueHandler) checkin(w http.ResponseWriter, r *http.Request) { 
    w.Header().Set("Content-Type", "application/json")
    queue := qh.PortQueue
    var port Request
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()

    err := decoder.Decode(&port)
    if err != nil {
	e := fmt.Sprintf("Error decoding request: %s", err)
	res, _ := json.Marshal(CheckinResponse{
	    Message: e,
	})
	w.Write(res)

	log.Printf("invalid request: %s\n", err)
	return
    }
    queue.PushBack(port.Port)
    log.Printf("returned port %d back to queue\n", port.Port)

    message := fmt.Sprintf("Port %d returned to queue", port.Port)
    res, _:= json.Marshal(CheckinResponse{
	Message: message,
    })

    w.Write(res)
}
