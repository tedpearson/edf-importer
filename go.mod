module github.com/tedpearson/edf-importer

go 1.21

require (
	github.com/influxdata/influxdb-client-go/v2 v2.13.0
	github.com/ishiikurisu/edf v1.0.1-0.20210127125851-80fff761e798
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	golang.org/x/net v0.21.0 // indirect
)

replace github.com/ishiikurisu/edf => github.com/tedpearson/edf v0.0.0-20240210051734-e040e9508b51