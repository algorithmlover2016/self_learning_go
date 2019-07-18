package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	fmt.Printf("Float64ToUint64 result: %v\n", ^uint64(0x0))
	fmt.Println(time.Second)
	fmt.Println(logrus.AllLevels)
	fmt.Printf("%T, %v\n", logrus.AllLevels, logrus.AllLevels)
}
