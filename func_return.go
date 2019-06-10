package main

import (
	"fmt"
)

func main() {
	// the _ is necessary
	if err, _ := fun1(); err == nil {
		fmt.Println("success withour second return value")
	}

	if err, message := fun1(); err == nil {
		fmt.Println(message)
	}

	if err, message := fun2(); err != nil {
		fmt.Println(message)
	}
	fmt.Println(fmt.Sprintf("print integer %d", 10))
}

func fun1() (error, string) {
	var err error = nil
	// err := nil // the definition and declare is not correct, because every type can be nil
	message := fmt.Sprintf("%s", "hello world")
	return err, message
}

func fun2() (error, string) {
	message := fmt.Sprintf("%s", "hello world")
	return fmt.Errorf(message), message
}
