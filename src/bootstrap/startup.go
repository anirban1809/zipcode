package bootstrap

import (
	"fmt"
	"os"
)

func Startup() {
	dir, err := os.Getwd()

	if err != nil {
		fmt.Println("Failed to get working directory")
	}

	fmt.Printf("Working directory: %s\n", dir)
}
