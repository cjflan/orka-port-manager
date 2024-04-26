package main

import "log"

func (pqh *PortQueueHandler) init(start int, length int) {
	for i := 0; i < length; i++ {
		pqh.PortQueue.PushBack(start + i)
	}
	log.Println("finished initilizating queue")
}
