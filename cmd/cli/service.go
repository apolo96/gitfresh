package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func isAgentRunning() (bool, error) {
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER, APP_AGENT_FILE)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error al leer el archivo PID: %v\n", err)
		return false, err
	}
	pidstr := strings.TrimSpace(string(content))
	if pidstr == "" {
		return false, err
	}
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		fmt.Println("Error during conversion")
		return false, err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = process.Signal(os.Signal(syscall.Signal(0)))
	if err != nil {
		return false, err
	}
	return true, nil
}

func saveAgentPID(pid int) error {
	pidStr := fmt.Sprint(pid)
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER, APP_AGENT_FILE)
	return os.WriteFile(path, []byte(pidStr), 0644)
}
