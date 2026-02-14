package tools

import (
	"os"
)

func CreateFile(fileName string) error {
	fd, err := os.Create(fileName)

	if err != nil {
		return err
	}

	fd.Close()
	return nil
}
