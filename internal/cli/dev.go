// Package cli implements FluxGo development commands.
package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	watchInterval = 400 * time.Millisecond
	stopTimeout   = 3 * time.Second
)

type fileState struct {
	modified int64
	size     int64
}

type application struct {
	command *exec.Cmd
	done    chan error
	binary  string
}

// Dev runs the application and rebuilds it when source files change.
func Dev() error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("enter project root: %w", err)
	}

	fmt.Printf("FluxGo dev server\n")
	fmt.Printf("project: %s\n", root)
	fmt.Printf("watching: .go, .gohtml, .env, go.mod, go.sum\n\n")

	state, err := scanProject(root)
	if err != nil {
		return err
	}

	var running *application
	running, err = rebuild(root, running)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initial build failed: %v\n", err)
	}

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		select {
		case <-signals:
			fmt.Println("\nstopping development server...")
			return stop(running)
		case <-ticker.C:
			if running != nil {
				select {
				case runErr := <-running.done:
					if runErr != nil {
						fmt.Fprintf(os.Stderr, "application stopped: %v\n", runErr)
					}
					_ = os.Remove(running.binary)
					running = nil
				default:
				}
			}

			nextState, scanErr := scanProject(root)
			if scanErr != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", scanErr)
				continue
			}
			if statesEqual(state, nextState) {
				continue
			}

			state = nextState
			fmt.Printf("\nchange detected at %s\n", time.Now().Format("15:04:05"))

			next, buildErr := rebuild(root, running)
			running = next
			if buildErr != nil {
				fmt.Fprintf(os.Stderr, "build failed: %v\n", buildErr)
				continue
			}
		}
	}
}

func rebuild(root string, current *application) (*application, error) {
	buildDirectory := filepath.Join(root, ".flux", "tmp")
	if err := os.MkdirAll(buildDirectory, 0o755); err != nil {
		return current, fmt.Errorf("create build directory: %w", err)
	}

	binary := filepath.Join(
		buildDirectory,
		fmt.Sprintf("app-%d%s", time.Now().UnixNano(), executableExtension()),
	)
	build := exec.Command("go", "build", "-o", binary, "./bootstrap")
	build.Dir = root
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	fmt.Println("building...")
	if err := build.Run(); err != nil {
		_ = os.Remove(binary)
		return current, err
	}

	if err := stop(current); err != nil {
		_ = os.Remove(binary)
		return current, err
	}

	next, err := start(root, binary)
	if err != nil {
		_ = os.Remove(binary)
		return nil, err
	}
	fmt.Println("application started")
	return next, nil
}

func start(root, binary string) (*application, error) {
	command := exec.Command(binary)
	command.Dir = root
	command.Env = os.Environ()
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
	}

	app := &application{
		command: command,
		done:    make(chan error, 1),
		binary:  binary,
	}
	go func() {
		app.done <- command.Wait()
	}()
	return app, nil
}

func stop(app *application) error {
	if app == nil {
		return nil
	}

	if app.command.Process != nil {
		_ = app.command.Process.Signal(os.Interrupt)
	}

	select {
	case <-app.done:
	case <-time.After(stopTimeout):
		if app.command.Process != nil {
			if err := app.command.Process.Kill(); err != nil && !isProcessFinished(err) {
				return fmt.Errorf("stop application: %w", err)
			}
		}
		<-app.done
	}

	_ = os.Remove(app.binary)
	return nil
}

func scanProject(root string) (map[string]fileState, error) {
	state := make(map[string]fileState)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if path != root && ignoredDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !watchedFile(entry.Name()) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		state[relative] = fileState{
			modified: info.ModTime().UnixNano(),
			size:     info.Size(),
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan project: %w", err)
	}
	return state, nil
}

func ignoredDirectory(name string) bool {
	switch name {
	case ".git", ".flux", "vendor", "node_modules":
		return true
	default:
		return false
	}
}

func watchedFile(name string) bool {
	return strings.HasSuffix(name, ".go") ||
		strings.HasSuffix(name, ".gohtml") ||
		name == ".env" ||
		name == "go.mod" ||
		name == "go.sum"
}

func statesEqual(left, right map[string]fileState) bool {
	if len(left) != len(right) {
		return false
	}
	for path, state := range left {
		if right[path] != state {
			return false
		}
	}
	return true
}

func findProjectRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("go.mod not found in this directory or its parents")
		}
		directory = parent
	}
}

func executableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func isProcessFinished(err error) bool {
	return errors.Is(err, os.ErrProcessDone)
}
