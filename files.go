package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type State struct {
	LastData time.Time
}

func readState(file string) State {
	def := func() State {
		return State{
			LastData: time.UnixMilli(0),
		}
	}
	f, err := os.ReadFile(file)
	if err != nil {
		return def()
	}
	var state State
	err = yaml.Unmarshal(f, &state)
	if err != nil {
		return def()
	}
	return state
}

func writeState(state State, file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Printf("failed to open state file for writing: %s\n", file)
		return
	}
	bytes, err := yaml.Marshal(state)
	if err != nil {
		fmt.Println("failed to marshal data")
		return
	}
	_, err = f.Write(bytes)
	if err != nil {
		fmt.Printf("failed to write state to: %s\n", file)
	}
}

func findFiles(dir string, lastUpdated time.Time) ([]string, error) {
	// get dirs
	dateDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	filteredFiles := make([]string, 0, 100)
	// BRP: flow/pressure
	// PLD: mask/press/leak/rep/tid/snbore/flowlim/etc
	// EVE: annotations
	regex := regexp.MustCompile(`^[^.].+(BRP|PLD|EVE).edf$`)
	for _, dateDir := range dateDirs {
		date, err := time.ParseInLocation("20060102", dateDir.Name(), time.Local)
		if err != nil || date.Before(lastUpdated.Add(-time.Hour*48)) || !dateDir.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(dir, dateDir.Name()))
		for _, file := range files {
			fileInfo, err := file.Info()
			if err != nil {
				panic(err)
			}
			if regex.MatchString(file.Name()) && fileInfo.Size() > 0 {
				filteredFiles = append(filteredFiles, filepath.Join(dir, dateDir.Name(), file.Name()))
			}
		}
	}
	return filteredFiles, nil
}

// RunWhenMediaInserted calls f when file becomes available. If file becomes unavailable and then available again,
// f will be called each time.
func RunWhenMediaInserted(file string, ctx context.Context, f func()) {
	fsFound := false
	fmt.Printf("Watching media path: %s\n", file)
	for {
		fileInfo, err := os.Stat(file)
		if fsFound {
			if err != nil {
				fsFound = false
				fmt.Printf("Media removed: %s\n", file)
			}
		} else {
			if err == nil && fileInfo.IsDir() {
				fsFound = true
				fmt.Printf("Media inserted: %s\n", file)
				f()
			}
		}
		select {
		case <-ctx.Done():
			fmt.Printf("Stopping watching path: %s", file)
			return
		case <-time.After(time.Second * 5):
			// continue to watch
		}
	}
}
