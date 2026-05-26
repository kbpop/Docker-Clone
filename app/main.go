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

func printPid(){
	pid := os.Getpid()
	fmt.Printf("Current Process ID: %d\n", pid)
}

func printProcs(){
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))	


}

// isolate filesystem 
func isolateFs(jailpath string, command string) {
	newpath := filepath.Join(jailpath, command)
	targetDir := filepath.Dir(newpath)
	os.MkdirAll(targetDir, 0755)

	srcFile, err := os.Open(command)
	if err!= nil {
		fmt.Printf("Err opening source command: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	destFile, err := os.Create(newpath)
	if err!= nil {
		fmt.Printf("Err creating destination file: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err!= nil {
		panic(err)
	}

	fileInfo, err := os.Stat(command)
	if err!= nil {
		fmt.Printf("Err reading permissions for %s: %v\n", command, err)
		os.Exit(1)
	}
	
	err = os.Chmod(newpath, fileInfo.Mode().Perm())
	if err!= nil {
		panic(err)
	}
}

// isolateProc mounts the virtual /proc filesystem inside the jail
func isolateProc() {
	procdir := "/proc"
	os.MkdirAll(procdir, 0755)

	src := "proc"
	target := procdir
	fstype := "proc"
	data := ""

	err := syscall.Mount(src, target, fstype, 0, data)
	if err!= nil {
		log.Fatalf("Mount failed: %v", err)
	}
}

// parentMode executes the host binary recursively within isolated namespaces
func parentMode() {
	args := os.Args[3:]
	childArgs := append(string{"child"}, args...)
	cmd := exec.Command("/proc/self/exe", childArgs...)	

	println("parent-commands: ", childArgs)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err!= nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
	
	os.Exit(0)
}

// childMode configures the localized jail, mounts /proc, and executes the user command
func childMode() {
	command := os.Args
	args := os.Args[3:]
	
	println("child-commands: ", args)
	
	jailpath := "/tmp/docker_jail"

	isolateFs(jailpath, command)

	// Declare root mount propagation as private to block host leakages
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err!= nil {
		log.Fatalf("Mount propagation error: %v", err)
	}

	if err := syscall.Chroot(jailpath); err!= nil {
		log.Fatalf("Chroot error: %v", err)
	}
	if err := syscall.Chdir("/"); err!= nil {
		log.Fatalf("Chdir error: %v", err)
	}

	isolateProc()

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err!= nil {	
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}	
	
	os.Exit(0)
}

func main() {
	hierarchy := os.Args

	if hierarchy == "child" {
		childMode()
		return
	}

	parentMode()
}