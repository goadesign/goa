package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("must pass a prefix arg to be ignored, to test it being passed")
		os.Exit(1)
	}
	cmd := exec.Command("protoc", os.Args[2:]...)
	fmt.Println(cmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	os.Exit(cmd.ProcessState.ExitCode())
}
