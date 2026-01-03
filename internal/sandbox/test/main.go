package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("WASM_SUCCESS: Args=%v", os.Args[1:])
}
