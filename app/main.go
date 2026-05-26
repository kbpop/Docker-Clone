package main

import (
	"fmt"
	"os"
	"os/exec"
	"io"
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

	// fmt.Printf("executing: %s", command)
	cmd := exec.Command(command, args...)

	// Give the child process the jailpath
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot: jailpath,
	}

	cmd.Stderr = os.Stderr // create error pipe for Go collection

	// create local files that the command needs	
	newpath := filepath.Join(jailpath, command)
	targetDir := filepath.Dir(newpath)
	// fmt.Printf("creating: %s", targetDir)
	os.MkdirAll(targetDir, 0755)

	// copy over bin now

	// file to read from in current local dir
	srcFile, err := os.Open(command)
	if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}
	defer srcFile.Close() // close file for later

	// file to write to
	destFile, err := os.Create(newpath)
	if err != nil {
		fmt.Printf("Err creating destination file: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close() // close file for later

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		panic(err)
	}

	// assign local file permissions to new file 
	fileInfo, err := os.Stat(command)
	err = os.Chmod(newpath, fileInfo.Mode().Perm())

	if err != nil {
		panic(err)
	}
	destFile.close()
	srcFile.close()

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
