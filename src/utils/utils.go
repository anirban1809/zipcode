package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"zipcode/src/config"
	"zipcode/src/tools"

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

func GetTool(path string, toolname string) (tools.Tool, error) {
	name := strings.ReplaceAll(toolname, "_tool", "")
	content, err := os.ReadFile(fmt.Sprintf("%s/%s/%s.json", path, name, name))

	if err != nil {
		return tools.Tool{}, errors.New("failed to read tool manifest")
	}

	var tool tools.Tool
	err = json.Unmarshal([]byte(content), &tool)

	if err != nil {
		return tools.Tool{}, errors.New("invalid tool manifest")
	}

	return tool, nil
}
