package main

import (
	"log"

	"github.com/gammazero/deque"
)

type PortQueueHandler struct {
    PortQueue *deque.Deque[int]
}

func (pqh *PortQueueHandler) init(start int, length int) {
	for i := 0; i < length; i++ {
		pqh.PortQueue.PushBack(start + i)
	}
	log.Println("finished initilizating queue")
}
