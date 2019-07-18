package main

import (
	"fmt"
)

type size int64
type str string
type stct struct {
	Name string
	Id   int64
}
type st stct

func PrintHello() {
	fmt.Println("hello world")
}

type itf interface {
	PrintStr(msg string)
}

func (str1 *str) PrintStr(msg string) {
	fmt.Println(string(*str1) + msg)
}

var printEx = PrintHello

func main() {
	var tmp1 size = 56
	fmt.Println("tmp1: ", tmp1)
	var str1 str = "fdsafasdfasf"
	fmt.Println("str1: ", str1)
	st1 := st{
		Name: "fdsaff",
		Id:   434232423,
	}
	fmt.Printf("st: %v\n", st1)
	printEx()
	PrintHello()
	str1.PrintStr("hello")

}
