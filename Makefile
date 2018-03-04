YAML=$(wildcard *.yaml)
JSON=$(YAML:.yaml=.json)
.SUFFIXES: .yaml .json
.yaml.json:
	yq read $< --tojson hooks |jq . > $@

all: $(JSON) run

run: run.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o $@ \
	-a -installsuffix cgo \
	-ldflags="-s -w"
