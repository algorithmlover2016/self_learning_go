package main

import (
    "fmt"
    "time"
)

func main() {
    buffer := make (chan int)
    go worker(buffer)
    for flag, index:= true, 0; flag; {
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
    ch := make(chan int, 4) // 0, 1, other length will lead to different result
    go worker(buffer)
    for i := 0; i < 10; i++ {
        select {
        case x := <- buffer:
            fmt.Println(x)
        case ch <- i: {
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