package main

import "fmt"

var phil0meals, phil1meals, phil2meals, phil3meals, phil4meals, phil5meals int

func main() {
	go fork0()
	go fork1()
	go fork2()
	go fork3()
	go fork4()
	//go fork5()
	go phil0()
	go phil1()
	go phil2()
	go phil3()
	go phil4()
	//go phil5()

	for true {
		if phil0meals >= 3 && phil1meals >= 3 && phil2meals >= 3 && phil3meals >= 3 && phil4meals >= 3 /*&& phil5meals >= 3*/ {
			break
		}
	}
	fmt.Println("slut, prut, finale:)")
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

//channels

var ch0_1 = make(chan int)
var ch1_2 = make(chan int)
var ch2_3 = make(chan int)
var ch3_4 = make(chan int)
var ch4_0 = make(chan int)

//var ch5_0 = make(chan int)

// all phils
func phil0() {
	var even, odd int
	fmt.Println("phil0 is Thinking??")
	for true {
		even = <-ch4_0
		if even == 1 {
			odd = <-ch0_1
			if even == 1 && odd == 1 {
				phil0meals++
				fmt.Println("phil0: nam nam....")
				ch4_0 <- 1
				ch0_1 <- 1
				fmt.Println("phil0 is Thinking??")
			}
		}
	}
}

func phil1() {
	var even, odd int
	fmt.Println("phil1 is Thinking??")
	for true {
		even = <-ch1_2
		if even == 1 {
			odd = <-ch0_1
			if even == 1 && odd == 1 {
				phil1meals++
				fmt.Println("Phil1: nam nam....")
				ch1_2 <- 1
				ch0_1 <- 1
				fmt.Println("phil1 is Thinking??")
			}
		}
	}
}
func phil2() {
	var even, odd int
	fmt.Println("phil2 is Thinking??")
	for true {
		even = <-ch1_2
		if even == 1 {
			odd = <-ch2_3
			if even == 1 && odd == 1 {
				phil2meals++
				fmt.Println("Phil2: nam nam....")
				ch1_2 <- 1
				ch2_3 <- 1
				fmt.Println("phil2 is Thinking??")
			}
		}
	}
}
func phil3() {
	var even, odd int
	fmt.Println("phil3 is Thinking??")
	for true {
		even = <-ch3_4
		if even == 1 {
			odd = <-ch2_3
			if even == 1 && odd == 1 {
				phil3meals++
				fmt.Println("Phil3: nam nam....")
				ch2_3 <- 1
				ch3_4 <- 1
				fmt.Println("phil3 is Thinking??")
			}
		}
	}
}
func phil4() {
	var even, odd int
	fmt.Println("phil4 is Thinking??")
	for true {
		even = <-ch3_4
		if even == 1 {
			odd = <-ch4_0
			if even == 1 && odd == 1 {
				phil4meals++
				fmt.Println("Phil4: nam nam....")
				ch4_0 <- 1
				ch3_4 <- 1
				fmt.Println("phil4 is Thinking??")
			}
		}
	}
}

/*func phil5() {
	var even, odd int
	fmt.Println("phil5 is Thinking??")
	for true {
		even = <-ch5_0
		if even == 1 {
			odd = <-ch4_5
			if even == 1 && odd == 1 {
				phil0meals++
				fmt.Println("phil5: nam nam....")
				ch5_0 <- 1
				ch4_5 <- 1
				fmt.Println("phil5 is Thinking??")
			}
		}
	}
}*/

// all forks
func fork0() {
	ch4_0 <- 1
	fmt.Println("Fork0 is free")
}
func fork1() {
	ch0_1 <- 1
	fmt.Println("Fork1 is free")
}
func fork2() {
	ch1_2 <- 1
	fmt.Println("Fork2 is free")
}
func fork3() {
	ch2_3 <- 1
	fmt.Println("Fork3 is free")
}
func fork4() {
	ch3_4 <- 1
	fmt.Println("Fork4 is free")
}

/*func fork5() {
	ch4_5 <- 1
	fmt.Println("Fork5 is free")
}*/
