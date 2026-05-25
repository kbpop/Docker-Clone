package main

import (
	"fmt"
	"os"
	"os/exec"
	"log"
	"syscall"
	"path/filepath"
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
	jailpath := "/tmp/docker_jail"

	os.MkdirAll(jailpath, 0755)

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]
	
	// isolate filesystem 
	// before isolation

	fmt.Println("executing: %s", command)
	cmd := exec.Command(command, args...)

	// Give the child process the jailpath
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot: jailpath,
	}

	cmd.Stderr = os.Stderr // create error pipe for Go collection

	// create local files that the command needs	
	newpath := filepath.Join(jailpath, command)
	targetDir := filepath.Dir(newpath)
	fmt.Println("creating: %s", targetDir)
	os.MkdirAll(targetDir, 0755)

	// copy over bin now

	// open file to copy over
	file, err := os.Open(command); err != nil {
		fmt.Printf("Err: %v", err)
	}
	// close file later
	defer file.Close()
	bytes, err := io.Copy(targetDir, command)
	if err != nil {
		panic(err)
	}

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
