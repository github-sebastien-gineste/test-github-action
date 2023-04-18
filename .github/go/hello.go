package main

import (
	"fmt"
)

func main() {
	myInput := "test"

	output := fmt.Sprintf("Hello %s", myInput)

	fmt.Println(fmt.Sprintf(`"{myOutput}={%s}" >> $GITHUB_OUTPUT`, output))
}

