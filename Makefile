YAML=$(wildcard *.yaml)
JSON=$(YAML:.yaml=.json)
.SUFFIXES: .yaml .json
.yaml.json:
	yq read $< --tojson hooks |jq . > $@

all: $(JSON)
