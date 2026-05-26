package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Container struct {
	Command string
	Args    []string
	RootFS  string
}

func (c *Container) Run() error {
	cmd := exec.Command(c.Command, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot:     c.RootFS,
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getegid(), Size: 1},
		},
	}

	return cmd.Run()
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// Setup isolated filesystem
	rootFSDir, err := setupRootFS(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up root filesystem: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(rootFSDir)

	container := &Container{
		Command: command,
		Args:    args,
		RootFS:  rootFSDir,
	}

	if err := container.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Container execution failed: %v\n", err)
		os.Exit(1)
	}
}

func setupRootFS(src string) (string, error) {
	dir, err := os.MkdirTemp("", "chroot")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	dest := filepath.Join(dir, src)

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return dir, fmt.Errorf("failed to create app directories: %w", err)
	}
	srcBytes, err := os.ReadFile(src)
	if err != nil {
		return dir, fmt.Errorf("failed to read source executable: %w", err)
	}

	if err := os.WriteFile(dest, srcBytes, 0755); err != nil {
		return dir, fmt.Errorf("failed to copy executable to rootfs: %w", err)
	}

	devDir := filepath.Join(dir, "dev")
	if err := os.MkdirAll(devDir, 0755); err != nil {
		return dir, fmt.Errorf("failed to create /dev directory: %w", err)
	}

	devNull := filepath.Join(devDir, "null")
	if err := os.WriteFile(devNull, []byte{}, 0644); err != nil {
		return dir, fmt.Errorf("failed to create /dev/null: %w", err)
	}

	return dir, nil
}
