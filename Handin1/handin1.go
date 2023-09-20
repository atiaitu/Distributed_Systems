// THIS PIECE OF CODE DOENS'T REACH A DEADLOCK BECUASE 
// THE FORKS CHANNEL "free" AND EVEN THOUGH BOTH PHILOSOPHORS CLOSE 
// TO THE FORK LISTEN IN THAT CHANNEL, ONLY ONE WILL RETREIVE 
// THE MESSAGE, THE OTHER IS STILL WAITING TO RECEIVE THE MESSAGE 
// "free", ONLY WHEN THE PHILOSOPHOR HAS ITS EVEN KNIFE 
// (NUMBERED 0-4), HE WILL TRY FOR THE ODD KNIFE, AND THE SAME 
// HAPPENS, UNTIL HE HAS BOTH HIS KNIFES AND HE SIGNALS TO BOTH 
// KNIFES THAT HE IS DONE USING THEM. THE FORKS LISTENS FOR THE 
// "done" MESSAGE FROM BOTH NEARBY PHILOSOPHORS, BUT BY USING 
// SELECT, STOPS LISTENING AFTER RECEIVING THE MESSAGE DONE FROM 
// ONE OF THEM, AND THEN AGAIN MESSAGING "free" TO THE CHANNEL 
// THAT BOTH PHILOSOPHORS CAN RECEIVE FROM

package main

import (
	"fmt"
)

var phil0meals, phil1meals, phil2meals, phil3meals, phil4meals, phil5meals int

func main() {
	go fork0()
	go fork1()
	go fork2()
	go fork3()
	go fork4()

	go phil0()
	go phil1()
	go phil2()
	go phil3()
	go phil4()

	for true {
		if phil0meals >= 3 && phil1meals >= 3 && phil2meals >= 3 && phil3meals >= 3 && phil4meals >= 3 {
			break
		}
	}
	fmt.Print(" phil0: ")
	fmt.Println(phil0meals)
	fmt.Print(" phil1: ")
	fmt.Println(phil1meals)
	fmt.Print(" phil2: ")
	fmt.Println(phil2meals)
	fmt.Print(" phil3: ")
	fmt.Println(phil3meals)
	fmt.Print(" phil4: ")
	fmt.Println(phil4meals)

}

// channels
var chP0_F0 = make(chan string)
var chP1_F0 = make(chan string)
var chF0_P0P1 = make(chan string)

var chP1_F1 = make(chan string)
var chP2_F1 = make(chan string)
var chF1_P1P2 = make(chan string)

var chP2_F2 = make(chan string)
var chP3_F2 = make(chan string)
var chF2_P2P3 = make(chan string)

var chP3_F3 = make(chan string)
var chP4_F3 = make(chan string)
var chF3_P3P4 = make(chan string)

var chP4_F4 = make(chan string)
var chP0_F4 = make(chan string)
var chF4_P4P0 = make(chan string)

func forks(message string, inUse bool, ch1 chan string, ch2 chan string, ch3 chan string) {
	if message == "done" {
		ch3 <- "free"
	}
}

func phil0() {
	var answer string
	fmt.Println("thinking")
	for true {
		answer = <-chF0_P0P1
		if answer == "free" {
			answer = <-chF4_P4P0
			fmt.Println(answer)
			if answer == "free" {
				phil0meals++

				fmt.Print("phil 0 is eating and has eaten: ")
				fmt.Println(phil0meals)
				fmt.Println("phil 0 is thinking again")
				chP0_F0 <- "done"
				chP0_F4 <- "done"
			}
		}
	}
}

func phil1() {
	var answer string
	fmt.Println("thinking")
	for true {
		answer = <-chF0_P0P1
		if answer == "free" {
			answer = <-chF1_P1P2
			fmt.Println(answer)
			if answer == "free" {
				phil1meals++

				fmt.Print("phil 1 is eating and has eaten: ")
				fmt.Println(phil1meals)
				fmt.Println("phil 1 is thinking again")
				chP1_F0 <- "done"
				chP1_F1 <- "done"
			}
		}
	}
}

func phil2() {
	var answer string
	fmt.Println("thinking")
	for true {
		answer = <-chF2_P2P3
		if answer == "free" {
			answer = <-chF1_P1P2
			fmt.Println(answer)
			if answer == "free" {
				phil2meals++

				fmt.Print("phil 2 is eating and has eaten: ")
				fmt.Println(phil2meals)
				fmt.Println("phil 2 is thinking again")
				chP2_F2 <- "done"
				chP2_F1 <- "done"
			}
		}
	}
}

func phil3() {
	var answer string
	fmt.Println("thinking")
	for true {
		answer = <-chF2_P2P3
		if answer == "free" {
			answer = <-chF3_P3P4
			fmt.Println(answer)
			if answer == "free" {
				phil3meals++

				fmt.Print("phil 3 is eating and has eaten: ")
				fmt.Println(phil3meals)
				fmt.Println("phil 3 is thinking again")
				chP3_F2 <- "done"
				chP3_F3 <- "done"
			}
		}
	}
}

func phil4() {
	var answer string
	fmt.Println("thinking")
	for true {
		answer = <-chF4_P4P0
		if answer == "free" {
			answer = <-chF3_P3P4
			fmt.Println(answer)
			if answer == "free" {
				phil4meals++

				fmt.Print("phil 4 is eating and has eaten: ")
				fmt.Println(phil4meals)
				fmt.Println("phil 4 is thinking again")
				chP4_F4 <- "done"
				chP4_F3 <- "done"
			}
		}
	}
}

// all forks
func fork0() {
	var inUse bool
	chF0_P0P1 <- "free"
	for true {
		select {
		case message := <-chP0_F0:
			forks(message, inUse, chP0_F0, chP1_F0, chF0_P0P1)
		case message := <-chP1_F0:
			forks(message, inUse, chP0_F0, chP1_F0, chF0_P0P1)
		}
	}
}

func fork1() {
	var inUse bool
	chF1_P1P2 <- "free"

	for true {
		select {
		case message := <-chP1_F1:
			forks(message, inUse, chP1_F1, chP2_F1, chF1_P1P2)
		case message := <-chP2_F1:
			forks(message, inUse, chP1_F1, chP2_F1, chF1_P1P2)
		}
	}
}

func fork2() {
	var inUse bool
	chF2_P2P3 <- "free"
	for true {
		select {
		case message := <-chP2_F2:
			forks(message, inUse, chP2_F2, chP3_F2, chF2_P2P3)
		case message := <-chP3_F2:
			forks(message, inUse, chP2_F2, chP3_F2, chF2_P2P3)
		}
	}
}

func fork3() {
	var inUse bool
	chF3_P3P4 <- "free"
	for true {
		select {
		case message := <-chP3_F3:
			forks(message, inUse, chP3_F3, chP4_F3, chF3_P3P4)
		case message := <-chP4_F3:
			forks(message, inUse, chP3_F3, chP4_F3, chF3_P3P4)
		}
	}
}

func fork4() {
	var inUse bool
	chF4_P4P0 <- "free"
	for true {
		select {
		case message := <-chP4_F4:
			forks(message, inUse, chP4_F4, chP0_F4, chF4_P4P0)
		case message := <-chP0_F4:
			forks(message, inUse, chP4_F4, chP0_F4, chF4_P4P0)
		}
	}
}
