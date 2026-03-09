package utils

import (
	"encoding/json"
	"fmt"
)

func PrintStruct[T any](value T) {
	result, _ := json.Marshal(value)
	fmt.Println(string(result))
}
