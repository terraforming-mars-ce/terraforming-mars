package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	serverProcess   *exec.Cmd
	restartDebounce = make(chan bool, 1)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/watch/main.go <command> [args...]")
		fmt.Println("Example: go run cmd/watch/main.go cmd/server/main.go")
		os.Exit(1)
	}

	command := os.Args[1:]

	// Start the debounce handler
	go handleRestart(command)

	// Start the server initially
	startServer(command)

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create watcher:", err)
	}
	defer watcher.Close()

	// Add directories to watch
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories
		if info.IsDir() {
			name := filepath.Base(path)
			if name == ".git" || name == "bin" || name == "node_modules" || name == "coverage.html" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		log.Fatal("Failed to add paths to watcher:", err)
	}

	fmt.Println("File watcher started...")
	fmt.Println("Watching for changes in Go files...")

	// Listen for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only watch .go files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			// Ignore certain events
			if event.Has(fsnotify.Chmod) {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				fmt.Printf("File changed: %s\n", event.Name)
				triggerRestart()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v\n", err)
		}
	}
}

func triggerRestart() {
	select {
	case restartDebounce <- true:
		// Successfully sent restart signal
	default:
		// Channel is full, restart already pending
	}
}

func handleRestart(command []string) {
	for range restartDebounce {
		// Debounce multiple rapid file changes
		time.Sleep(300 * time.Millisecond)

		// Drain any additional restart signals
		for {
			select {
			case <-restartDebounce:
				continue
			default:
				goto restart
			}
		}

	restart:
		stopServer()
		startServer(command)
	}
}

func startServer(command []string) {
	fmt.Println("Starting server...")

	// For single files, use go run directly on the file
	if len(command) == 1 {
		serverProcess = exec.Command("go", "run", command[0])
	} else {
		// For multiple files, build the go run command
		args := append([]string{"run"}, command...)
		serverProcess = exec.Command("go", args...)
	}

	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr

	err := serverProcess.Start()
	if err != nil {
		log.Printf("Failed to start server: %v\n", err)
		return
	}

	fmt.Printf("Server started (PID: %d)\n", serverProcess.Process.Pid)
}

func stopServer() {
	if serverProcess != nil && serverProcess.Process != nil {
		fmt.Printf("Stopping server (PID: %d)...\n", serverProcess.Process.Pid)

		// Try graceful shutdown first
		serverProcess.Process.Signal(os.Interrupt)

		// Wait a bit for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- serverProcess.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(2 * time.Second):
			// Force kill if graceful shutdown takes too long
			fmt.Println("Graceful shutdown timeout, force killing...")
			serverProcess.Process.Kill()
			<-done // Wait for process to actually exit
		}

		serverProcess = nil
	}
}
