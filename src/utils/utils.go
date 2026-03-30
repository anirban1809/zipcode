package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"zipcode/src/config"

	"golang.org/x/term"
)

func PrintStruct(value any) {
	result, _ := json.Marshal(value)
	fmt.Println(string(result) + "\n")
}

func isStruct(v any) bool {
	// If the value is a pointer, you must dereference it using .Elem()
	// before checking if it's a struct.
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Struct
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

func LogValue(value any) {
	if !config.HEADLESS {
		return
	}

	if isStruct(value) {
		PrintStruct(value)
		return
	}

	fmt.Println(value)

}

func Log(a ...any) {
	if config.HEADLESS {
		fmt.Println(a...)
	}
}
