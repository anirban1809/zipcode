package agent

import "fmt"

type Runtime struct {
}

func (r Runtime) Submit(prompt string) {
	fmt.Println("Running: ", prompt)
}
