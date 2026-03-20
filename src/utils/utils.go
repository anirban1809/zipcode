package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"zipcode/src/config"

	"golang.org/x/term"
)

func PrintStruct[T any](value T) {
	result, _ := json.Marshal(value)
	fmt.Println(string(result) + "\n")
}

func GetTerminalSize() (int, int, error) {
	fd := (os.Stdout.Fd())
	width, height, err := term.GetSize(int(fd))
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

func FlexGap(totalWidth int, subWidth int) string {
	gap := totalWidth - subWidth
	gapText := ""

	for range gap {
		gapText += " "
	}

	return gapText
}

func Log(a ...any) {
	if config.HEADLESS {
		fmt.Println(a...)
	}
}
