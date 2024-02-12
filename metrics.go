package main

import (
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxConfig is the configuration for Influx/VictoriaMetrics.
type InfluxConfig struct {
	Host      string
	AuthToken string `yaml:"auth_token"`
	Org       string
	Bucket    string
}

type InfluxWriter struct {
	client influxdb2.Client
	api    api.WriteAPI
}

func NewInfluxWriter(config InfluxConfig) InfluxWriter {
	client := influxdb2.NewClient(config.Host, config.AuthToken)
	return InfluxWriter{
		client: client,
		api:    client.WriteAPI(config.Org, config.Bucket),
	}
}

func (i InfluxWriter) Close() {
	i.client.Close()
}

func (i InfluxWriter) WriteData(annotations map[string]*Annotation, metrics []Metric) {
	i.writeAnnotations(annotations)
	i.writeMetrics(metrics)
}

func (i InfluxWriter) writeAnnotations(annotations map[string]*Annotation) {
	for _, annotation := range annotations {
		for _, event := range annotation.Events {
			i.api.WritePoint(influxdb2.NewPointWithMeasurement("cpap").
				AddTag("event", annotation.Name).
				AddField("annotation", 0).
				SetTime(event.Start.Add(-time.Second)))
			i.api.WritePoint(influxdb2.NewPointWithMeasurement("cpap").
				AddTag("event", annotation.Name).
				AddField("annotation", 1).
				SetTime(event.Start))
			i.api.WritePoint(influxdb2.NewPointWithMeasurement("cpap").
				AddTag("event", annotation.Name).
				AddField("annotation", 0).
				SetTime(event.End))
		}
	}
}

func (i InfluxWriter) writeMetrics(metrics []Metric) {
	for _, metric := range metrics {
		for _, point := range metric.Data {
			i.api.WritePoint(influxdb2.NewPointWithMeasurement("cpap").
				AddField(metric.Name, point.Datum).
				SetTime(point.Time))
		}
	}
}
