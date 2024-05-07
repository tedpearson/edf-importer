package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
)

var (
	version   = "development"
	goVersion = "unknown"
	buildDate = "unknown"
)

func main() {
	path := flag.String("path", "/Volumes/NO NAME/DATALOG", "Path to data directory")
	configFile := flag.String("config", "edf-importer.yaml", "Config file")
	stateFile := flag.String("state-file", "edf-importer.state.yaml", "State file")
	dryRun := flag.Bool("dry-run", false, "Don't insert into the database")
	versionFlag := flag.Bool("v", false, "Show version and exit")
	flag.Parse()
	log.Printf("edf-importer version %s built on %s with %s\n", version, buildDate, goVersion)
	if *versionFlag {
		os.Exit(0)
	}

	influxConfig := readConfig(*configFile)
	influxWriter := NewInfluxWriter(influxConfig)
	state := readState(*stateFile)

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	if *dryRun {
		log.Println("Dry run - no data will be inserted into the database.")
		importData(*path, &state, true, *stateFile, influxWriter, ctx)
		log.Println("Dry run - no data was inserted into the database.")
	} else {
		RunWhenMediaInserted(*path, ctx, func() {
			importData(*path, &state, false, *stateFile, influxWriter, ctx)
		})
	}
	influxWriter.Close()
}

func importData(path string, state *State, dryRun bool, stateFile string, influxWriter InfluxWriter, ctx context.Context) {
	files, err := findFiles(path, state.LastData)
	if err != nil {
		panic(err)
	}
	newLastData := state.LastData
	metricCount := 0
	annotationCount := 0
	metricFileCount := 0
	annotationFileCount := 0
	for _, file := range files {
		select {
		case <-ctx.Done():
			log.Println("Stop requested, not parsing any more files")
		default:
			// continue running
		}
		metrics, annotations, lastData, err := parseFile(file, state.LastData)
		if err != nil {
			log.Printf("Error parsing %s: %s\n", file, err)
			continue
		}
		metricCount += sumMetrics(metrics)
		annotationCount += sumAnnotations(annotations)
		if metricCount > 0 {
			metricFileCount++
		}
		if annotationCount > 0 {
			annotationFileCount++
		}
		if lastData.After(newLastData) {
			newLastData = lastData
		}
		if !dryRun {
			influxWriter.WriteData(annotations, metrics)
		}
	}
	state.LastData = newLastData
	if !dryRun {
		writeState(*state, stateFile)
	}
	log.Printf("\nTotal new data found: %s metric points in %s files, and %s annotation points in %s files.\n",
		humanize.Comma(int64(metricCount)), humanize.Comma(int64(metricFileCount)),
		humanize.Comma(int64(annotationCount)), humanize.Comma(int64(annotationFileCount)))
}

func readConfig(configFile string) InfluxConfig {
	cf, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file %s: %s", configFile, err)
	}
	var config InfluxConfig
	err = yaml.Unmarshal(cf, &config)
	if err != nil {
		log.Fatalf("Error loading config from %s: %s", configFile, err)
	}
	return config
}

func handleSignals(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	_ = <-c
	log.Println("Shutdown requested...")
	cancel()
}
