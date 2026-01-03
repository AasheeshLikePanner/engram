package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("Current System Time: %s", time.Now().Format(time.RFC3339))
}
