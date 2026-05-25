package main

import (
	"fmt"
	"os"
	"os/exec"
	"log"
	"syscall"
)

// Ensures gofmt doesn't remove the imports above (feel free to remove this!)
var _ = os.Args
var _ = exec.Command

func readDir(){
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Println("[DIR]", entry.Name())
		} else {
			fmt.Println("[FILE]", entry.Name())
		}
	}
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]
	

	// isolate filesystem 

	// before isolation
	readDir()

	// isolation step
	newRoot = "/temp"

	// create and change to directory 
	if err := syscall.Chdir(newRoot); err != nil {
		fmt.Println("Chdir error: %v", err)
	}

	// Create chroot jail
	if err := syscall.Chroot(newRoot); err != nil {
		fmt.Println("Chroot error: %v", err)
	}

	// change dir to "/" of new root
	if err := syscall.Chdir(newRoot); err != nil {
		fmt.Println("Chdir error: %v", err)
	}

	// after isolation
	readDir()


	cmd := exec.Command(command, args...)

	cmd.Stderr = os.Stderr // create error pipe for Go collection

	output, err := cmd.Output()
	if err != nil {	
		fmt.Printf("Err: %v", err)

		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}	
	
	fmt.Print(string(output))
	os.Exit(0)
}
