package main

import (
	"fmt"
	"time"
)

func main() {
	go client()
	go server()

	for true {
		if serverStatus == "done" {
			break
		}
	}
}

var serverStatus string

var syn = make(chan int, 2)

func client() {
	syn <- 1
	for true {
		select {
		case ack := <-syn:
			fmt.Println("C: message received")
			select {
			case seq := <-syn:
				if seq == 2 || ack == 2 {
					if seq == 2 {
						syn <- seq
						ack++
						syn <- ack
					} else {
						syn <- ack
						seq++
						syn <- seq
					}
					fmt.Println("Job finished, yubi")
					break
				}
				//fmt.Println("Job wiwiwiw finished, yubi")
			case <-time.After(2 * time.Second):
				fmt.Println("C: message failed, timeout reached")
			}
		case <-time.After(2 * time.Second):
			fmt.Println("C: second message failed, timeout reached")
		}
	}
}

func server() {
	for true {
		select {
		case ack := <-syn:
			var holder int = ack
			holder++
			syn <- holder
			syn <- 10
			fmt.Println("S: Message received!")

			select {
			case ack := <-syn:
				fmt.Println("S: Message received!")
				select {
				case seq := <-syn:
					fmt.Println("S: Second message received!")
					if seq == 11 || ack == 11 {
						fmt.Println("3-way handshake has been done!")
						break
					} else {
						fmt.Println("You fucking piece of shit")
					}

				case <-time.After(2 * time.Second):
					fmt.Println("S: Second message failed, timeout reached")
				}

			case <-time.After(2 * time.Second):
				fmt.Println("S: message failed, timeout reached")
			}

		case <-time.After(2 * time.Second):
			fmt.Println("S: Message failed, timeout reached")
		}
	}
}
