package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()
	args := flag.Args()
	fmt.Println("vim-go")
	var s byte
	s = 'a'
	if len(os.Args) > 1 {
		s = byte(os.Args[1][0])
	} else {
		fmt.Printf("args: %v\n", os.Args[0])
	}
	fmt.Printf("args: %v\n", args)
	switch s {
	case 'a':
		fmt.Println("The integer was <= 4")
		fallthrough
	case 'b':
		fmt.Println("The integer was <= 5")
		fallthrough
	case 'c':
		fmt.Println("The integer was <= 6")
		fallthrough
	case 'd':
		fmt.Println("The integer was <= 7")
	default:
		fmt.Println("default case")
	}
}
