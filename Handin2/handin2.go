package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	go myClient(rand.Intn(100))
	go myServer(rand.Intn(100))

	for true {
		if serverStatus == "done" {
			break
		}
	}
}

var serverStatus string

var syn = make(chan int)
var ack = make(chan int)

func myClient(code int) {
	syn <- code
	for true {
		select {
		case seq := <-syn:
			fmt.Println("C: Message received!")
			fmt.Println(seq)
			select {
			case ack_seq := <-ack:
				fmt.Println("C:Second message received!")
				fmt.Println(ack_seq)
				if seq == code+1 {
					syn <- seq
					ack <- ack_seq + 1
				} else {
					fmt.Println("C: Wrong Message!!!")
				}
			case <-time.After(2 * time.Second):
				fmt.Println("C: Second message failed, timeout reached")
			}
		case <-time.After(2 * time.Second):
			fmt.Println("C: Message failed, timeout reached")
		}
	}
}

func myServer(code int) {
	var holder int = 0
	for true {
		select {
		case seq := <-syn:
			fmt.Println("S: Message received")
			fmt.Println(seq)
			holder = seq + 1
			syn <- seq + 1
			ack <- code
		case <-time.After(2 * time.Second):
			fmt.Println("S: Message failed, timeout reached")
		}
		select {
		case seq := <-syn:
			fmt.Println("S: Second message received")
			fmt.Println(seq)
			select {
			case ack_seq := <-ack:
				fmt.Println("S: Second value message received")
				fmt.Println(ack_seq)
				if seq == holder && ack_seq == code+1 {
					fmt.Println("S: WE ARE DONE!")
					serverStatus = "done"
				} else {
					fmt.Println("S: Wrong message, pulling out...")
				}
			case <-time.After(2 * time.Second):
				fmt.Println("S: Wrong second value!")
			}
		case <-time.After(2 * time.Second):
			fmt.Println("S: Wrong value!")
		}
	}
}
