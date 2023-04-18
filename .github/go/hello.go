package main

import (
	"fmt"
	"os"
)

func main() {
	myInput := "test"

	output := fmt.Sprintf("Hello %s", myInput)

	fmt.Println(fmt.Sprintf(`::set-output name=myOutput::%s`, output))
}