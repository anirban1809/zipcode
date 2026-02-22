package tools

import (
	"os"
)

func CreateFile(fileName string) error {
	fd, err := os.Create(fileName)
	defer fd.Close()
	if err != nil {
		return err
	}

	return nil
}
