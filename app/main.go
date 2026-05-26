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
func isolateFs(jailpath string, command string){

	// create local files that the command needs	
	newpath := filepath.Join(jailpath, command)
	targetDir := filepath.Dir(newpath)
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

	// close files before moving on
	destFile.Close()
	srcFile.Close()
}

func isolateProc(){

	// create where proc filesystem will live
	procdir := "/proc"
	os.MkdirAll(procdir, 0755)

	src := "proc" // no actual hardware associated so dummy name
	target := procdir
	fstype := "proc" // create process filesystem
	// flags  := 0 // default options
	data   := ""// doesn't require any extra options

	err := syscall.Mount(src, target, fstype, 0, data)
	if err != nil {
		log.Fatalf("Mount failed: %v", err)
	}
}

func parentMode(){
	args := os.Args[3:]
	childArgs := append([]string{"child"}, os.Args[2],  args...)
	cmd := exec.Command("/proc/self/exe", childArgs...)	

	println("parent-commands: ", childArgs)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
	
	os.Exit(0)
}

func childMode(){
	command := os.Args[2]
	args := os.Args[3:]
	
	println("child-commands: ", args)
	// create command executable

	// isolate filesystem 
	jailpath := "/tmp/docker_jail"

	isolateFs(jailpath, command)

	// create Chroot manually
	if err := syscall.Chroot(jailpath); err != nil {
		log.Fatalf("Chroot error: %v", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		log.Fatalf("Chdir error: %v", err)
	}

	isolateProc()

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {	
		 // fmt.Printf("Err: %v", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}	
	
	os.Exit(0)
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	hierarchy := os.Args[1]

	if(hierarchy == "child"){
		childMode()
		return
	}

	parentMode()
}
