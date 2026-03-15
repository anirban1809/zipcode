package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/term"
)

func PrintStruct[T any](value T) {
	result, _ := json.Marshal(value)
	fmt.Println(string(result))
}

func GetTerminalSize() (int, int, error) {
	fd := (os.Stdout.Fd())
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}
