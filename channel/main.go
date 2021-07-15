package main

import (
	"fmt"
	"runtime"
	"time"
)

// reference to https://www.jianshu.com/p/245d9cbfe132
func producer(out chan<- int) {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)
	for i := 0; i < 10; i++ {
		out <- i * i
	}
	close(out)
	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func consumer(in <-chan int) {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)
	for num := range in {
		fmt.Println("num = ", num)
	}
	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func producerAndConsumer() {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)
	c := make(chan int)
	go producer(c)
	consumer(c)
	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func unbufferedDualChannelClose() {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)

	c := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			c <- i
		}
		close(c)
	}()

	for {
		if data, ok := <-c; ok {
			fmt.Println(data)
		} else {
			break
		}
	}

	cc := make(chan int)
	go func() {
		for i := 0; i < 5; i++ {
			cc <- i
		}
		close(cc)
	}()

	for data := range cc {
		fmt.Println(data)
	}

	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func bufferedDualChannel() {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)

	c := make(chan int, 3)

	//内置函数 len 返回未被读取的缓冲元素数量， cap 返回缓冲区大小
	fmt.Printf("len(c)=%d, cap(c)=%d\n", len(c), cap(c))

	go func() {
		defer fmt.Println("sub goroutine is over")

		for i := 0; i < 3; i++ {
			c <- i
			fmt.Printf("sub goroutine is running [%d]: len(c)=%d, cap(c)=%d\n", i, len(c), cap(c))
		}
	}()

	time.Sleep(2 * time.Second) //延时2s
	for i := 0; i < 3; i++ {
		num := <-c //从c中接收数据，并赋值给num
		fmt.Println("num = ", num)
	}
	fmt.Println("main routine is over")
	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func unbufferedDualChannel() {
	fName, file, line := runFuncName()
	fmt.Printf("call %s:%d function [%s] begin\n", file, line, fName)
	c := make(chan int)
	go func() {
		defer fmt.Println("sub-goroutine is over")
		fmt.Println("sub-goroutine is starting")
		c <- 666
	}()
	num := <-c
	fmt.Println("num = ", num)
	fmt.Println("main process has done")
	fmt.Printf("call %s:%d function [%s] is finished\n", file, line, fName)
}

func runFuncName() (string, string, int) {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	return f.Name(), file, line
}

func main() {
	unbufferedDualChannel()
	bufferedDualChannel()
	unbufferedDualChannelClose()
	producerAndConsumer()
}
