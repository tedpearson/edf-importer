package main

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ishiikurisu/edf"
)

type Metric struct {
	Name string
	Data []Point
}

type Point struct {
	Time  time.Time
	Datum float64
}

type Annotation struct {
	Name   string
	Events []Event
}

type Event struct {
	Start time.Time
	End   time.Time
}

func parseFile(file string, lastData time.Time) (metrics []Metric, annotations map[string]*Annotation, lastPointTime time.Time, e error) {
	log.Printf("Parsing %s\n", file)
	data := edf.ReadFile(file)
	samplingStart, err := time.ParseInLocation("02.01.06 15.04.05", data.Header["startdate"]+" "+data.Header["starttime"], time.Local)
	if err != nil {
		e = err
		return
	}
	annotationText := data.WriteNotes()
	if annotationText != "" {
		annotations = parseAnnotations(annotationText, samplingStart, lastData)
	}
	sampleRate := time.Millisecond * time.Duration(int(data.GetDuration()*1000)/data.GetSampling())
	lastPointTime = samplingStart.Add(sampleRate * time.Duration(len(data.PhysicalRecords[0])))
	if lastPointTime.After(lastData) {
		metrics = parseMetrics(data, samplingStart, sampleRate)
	}
	duration := lastPointTime.Sub(samplingStart)
	metricSum := sumMetrics(metrics)
	if metricSum > 0 {
		log.Printf("Found %s new points in %d metrics spanning %s (%s - %s)\n",
			humanize.Comma(int64(metricSum)), len(metrics), duration, samplingStart.Format(time.DateTime), lastPointTime.Format(time.DateTime))
	}
	annotationSum := sumAnnotations(annotations)
	if annotationSum > 0 {
		log.Printf("Found %s new events in %d annotations\n", humanize.Comma(int64(annotationSum)), len(annotations))
	}
	return
}

func sumMetrics(metrics []Metric) (sum int) {
	for _, metric := range metrics {
		sum += len(metric.Data)
	}
	return
}

func sumAnnotations(annotations map[string]*Annotation) (sum int) {
	for _, annotation := range annotations {
		sum += len(annotation.Events)
	}
	return
}

var annotationRE = regexp.MustCompile(`^\+([\d.]+)\s([\d.]+)\s(.+)\s+`)

func parseMetrics(data edf.Edf, samplingStart time.Time, sampleRate time.Duration) []Metric {
	metrics := make([]Metric, 0, 10)
	for i, series := range data.PhysicalRecords {
		name := strings.ReplaceAll(strings.TrimSpace(data.GetLabels()[i]), ".", "_")
		if name == "EDF Annotations" || name == "Crc16" {
			continue
		}
		points := make([]Point, 0, len(series))
		for j, value := range series {
			points = append(points, Point{
				Time:  samplingStart.Add(sampleRate * time.Duration(j)),
				Datum: math.Round(value*1000) / 1000,
			})
		}
		metric := Metric{
			Name: name,
			Data: points,
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func parseAnnotations(annotations string, samplingStart time.Time, lastData time.Time) map[string]*Annotation {
	// grab the timestamp and the string.
	parts := strings.Split(annotations, "\n")
	out := make(map[string]*Annotation)
	for _, part := range parts {
		match := annotationRE.FindStringSubmatch(part)
		if match == nil || len(match) < 4 {
			continue
		}
		onset, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			continue
		}
		duration, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		name := match[3]
		if name == "Recording starts" {
			continue
		}
		as, ok := out[name]
		if !ok {
			as = &Annotation{
				Name:   name,
				Events: make([]Event, 0, 10),
			}
			out[name] = as
		}
		// note: at least on a ResMed device, "onset" apparently refers to the end of the event,
		//       and the duration refers to how many seconds earlier the event started.
		//       I am not sure that's the actual definition of these fields in the EDF+ spec.
		endTime := samplingStart.Add(time.Millisecond * time.Duration(onset*1_000))
		if duration == 0 {
			// some events have no duration. defaulting them to 10s.
			duration = 10
		}
		onsetTime := endTime.Add(-time.Millisecond * time.Duration(duration*1_000))
		if endTime.After(lastData) {
			as.Events = append(as.Events, Event{
				Start: onsetTime,
				End:   endTime,
			})
		}
	}
	return out
}
