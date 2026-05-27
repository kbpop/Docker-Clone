package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"net/http"
	"encoding/json"
)

type RespAuth struct {
	Token    string    `json:"token"`
	Expires_in int `json:"expires_in"`
	Issued_at string `json:"issued_at"`
}

func getToken() string {
	url := "https://auth.docker.io/token"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to pull data: %v", err)
	}
	defer resp.Body.Close()

	var respAuth RespAuth
	if err := json.NewDecoder(resp.Body).Decode(&respAuth); err != nil {
    	log.Fatal(err)
	}
	return respAuth.Token
}

func authenticationDance(){

	// 1. Get a bearer token for the repository.
	token := getToken()

	// 2. Get the image manifest.

	// 3. If the response in the previous step is a multi-architecture manifest list, you must do the following:
	// 	o Parse the manifests[] array to locate the digest for your target platform (e.g., linux/amd64).
	// 	o Get the image manifest using the located digest.

	// 4. Check if the blob exists before downloading. The client should send a HEAD request for each layer digest.

	// 5. Download each layer blob using the digest obtained from the manifest. The client should send a GET request for each layer digest.
}

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

	// pull the image files in

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

	println("arg0 %s", args[0])
	println("arg1 %s", args[1])
	println("arg2 %s", args[2])
	println("arg3 %s", args[3])
	// authenticationDance()	
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