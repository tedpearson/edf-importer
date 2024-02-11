package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	version   = "development"
	goVersion = "unknown"
	buildDate = "unknown"
)

func main() {
	path := flag.String("path", "/Volumes/NO NAME/DATALOG", "Path to data directory")
	configFile := flag.String("config", "sdf-importer.yaml", "Config file")
	stateFile := flag.String("state-file", "edf-importer.state.yaml", "State file")
	versionFlag := flag.Bool("v", false, "Show version and exit")
	flag.Parse()
	fmt.Printf("edf-importer version %s built on %s with %s\n", version, buildDate, goVersion)
	if *versionFlag {
		os.Exit(0)
	}

	influxConfig := readConfig(*configFile)
	influxWriter := NewInfluxWriter(influxConfig)
	state := readState(*stateFile)
	files, err := findFiles(*path, state.LastData)
	if err != nil {
		panic(err)
	}
	newLastData := state.LastData
	for _, file := range files {
		metrics, annotations, err := parseFile(file, state.LastData)
		if len(metrics) > 0 && len(metrics[0].Data) > 0 {
			lastData := metrics[0].Data[len(metrics[0].Data)-1].Time
			if lastData.After(newLastData) {
				// update state if needed
				newLastData = lastData
			}
		}
		if err != nil {
			fmt.Printf("Error parsing %s: %s\n", file, err)
			continue
		}
		influxWriter.WriteData(annotations, metrics)
	}
	state.LastData = newLastData
	writeState(state, *stateFile)
	influxWriter.Close()
}

func readConfig(configFile string) InfluxConfig {
	cf, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Error reading config file %s: %s", configFile, err))
	}
	var config InfluxConfig
	err = yaml.Unmarshal(cf, &config)
	if err != nil {
		panic(fmt.Sprintf("Error loading config from %s: %s", configFile, err))
	}
	return config
}

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
		fmt.Printf("failed to write state to: %s", file)
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
			if regex.MatchString(file.Name()) {
				filteredFiles = append(filteredFiles, filepath.Join(dir, dateDir.Name(), file.Name()))
			}
		}
	}
	return filteredFiles, nil
}
