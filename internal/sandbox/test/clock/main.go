package main

import (
	"fmt"
	"time"
)

func main() {
	// The sandbox will pass the tool name as os.Args[0] and input as os.Args[1]
	// But here we just return the current time regardless.
	fmt.Print(time.Now().Format("2006-01-02 15:04:05"))
}
