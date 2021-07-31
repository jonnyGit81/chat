package main

import (
	"log"
	"os"
	"os/exec"
)

// The os/exec package gives us everything we need to work with to run external programs or commands from within Go programs.
// First, our cmdChain slice contains *exec.Cmd commands in the order in which we want to join them together.
var cmdChain = []*exec.Cmd{
	exec.Command("lib/synonyms"),
	exec.Command("lib/sprinkle"),
	exec.Command("lib/coolify"),
	exec.Command("lib/domainify"),
	exec.Command("lib/available"),
}

func main() {
	cmdChain[0].Stdin = os.Stdin
	cmdChain[len(cmdChain)-1].Stdout = os.Stdout
	for i := 0; i < len(cmdChain)-1; i++ {
		thisCmd := cmdChain[i]
		nextCmd := cmdChain[i+1]
		stdout, err := thisCmd.StdoutPipe()
		if err != nil {
			log.Fatalln(err)
		}
		nextCmd.Stdin = stdout
	}

	// We then iterate over each command calling the Start method, which runs the program in the background (as opposed to the Run method,
	// which will block our code until the subprogram exists which would be no good since we will have to run five programs at the same time).
	// If anything goes wrong, we bail with log.Fatalln; however, if the program starts successfully, we defer a call to kill the process.
	// This helps us ensure the subprograms exit when our main function exits, which will be when the domainfinder program ends.
	for _, cmd := range cmdChain {
		if err := cmd.Start(); err != nil {
			log.Fatalln(err)
		} else {
			defer cmd.Process.Kill()
		}
	}

	// Once all the programs start running, we iterate over every command again and wait for it to finish.
	// This is to ensure that domainfinder doesn't exit early and kill off all the subprograms too soon.
	for _, cmd := range cmdChain {
		if err := cmd.Wait(); err != nil {
			log.Fatalln(err)
		}
	}
}
