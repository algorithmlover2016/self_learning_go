package main

import (
    "fmt"
    "time"
)

func main() {
    buffer := make (chan int)
    defer close(buffer)
    go worker(buffer)
    for flag, index := true, 0; flag; {
        select {
        case num := <- buffer : {
            fmt.Println("buffer recieve:", num)
        }
        default: {
            fmt.Println("Hello word", "index:", index)
            index++
            flag = 100 > index
            time.Sleep(time.Microsecond * 200)
        }
        }
    }

    ch := make(chan int) // 0, 1, other length will lead to different result
    fmt.Println("zero length channel:", len(ch))
    // fmt.Println("read zero length channel:", <- ch)
    // ch <- 1
    // fmt.Println("zero length channel:", <- ch)
    close(ch)

    ch1 := make(chan int, 4) // 0, 1, other length will lead to different result
    defer close(ch1)
    go worker(buffer)
    for i := 0; i < 10; i++ {
        select {
        case x := <- ch:
            fmt.Println("invalid ch:", x)
        case x := <- ch1:
            fmt.Println("valid ch1:", x)
        case x := <- buffer:
            fmt.Println("valid buffer:", x)
        case ch1 <- i: {
            fmt.Println("send data %d into channel", i)
        }
        default :
            fmt.Println("do nothing")
        }
    }
}

func worker(buffer chan<- int) {
    for i := 0; i < 10; i++ {
        buffer <- i * 10
        time.Sleep(time.Microsecond * 500)
    }
}