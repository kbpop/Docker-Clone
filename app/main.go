package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func chrootSetup(entry string) string {
	chrootDir, err := os.MkdirTemp("", "chroot")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v\n", err)
	}

	err = os.Chdir(chrootDir)
	if err != nil {
		log.Fatalf("Failed to change directory: %v\n", err)
	}

	err = os.MkdirAll("usr/local/bin", 0o755)
	if err != nil {
		log.Fatalf("Failed creating binary dir: %v\n", err)
	}

	src, err := os.Open(entry)
	if err != nil {
		log.Fatalf("Failed copying file: %v\n", err)
	}

	info, err := src.Stat()
	if err != nil {
		log.Fatalf("Failed getting file info: %v\n", err)
	}

	defer src.Close()

	dest, err := os.OpenFile(fmt.Sprintf("%s%s", chrootDir, entry), os.O_CREATE|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		log.Fatalf("Failed creating file: %v\n", err)
	}

	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		log.Fatalf("Failed copying program to chroot: %v\n", err)
	}

	return chrootDir
}

func run(entry string, args []string) {
	chrootDir := chrootSetup(entry)

	defer os.RemoveAll(chrootDir)

	cmd := exec.Command(entry, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = "/"

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot: chrootDir,
		// Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	if args[0] == "mypid" {
		fmt.Println(1)
		return
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Failed to run: %v: %v\n", entry, err)
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("Here Failed to run: %v: %v\n", entry, err)
	}
}

func main() {
	usage := fmt.Sprintf("Usage: %s <cmd> <img> <entry> <args>", os.Args[0])
	if len(os.Args) < 4 {
		log.Fatalln(usage)
	}

	cmd := os.Args[1]
	_ = os.Args[2]

	entry := os.Args[3]

	switch cmd {

	case "run":
		run(entry, os.Args[4:])

	default:
		log.Fatalf("Invalid command: %s", cmd)
	}
}