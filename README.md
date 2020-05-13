# json-schema-docs

[![License](https://img.shields.io/github/license/grafana/json-schema-docs)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/grafana/json-schema-docs)](https://goreportcard.com/report/github.com/grafana/json-schema-docs)

A simple JSON Schema to Markdown generator.

This generator doesn't attempt to support the full JSON Schema specification. Instead, it's designed with the rationale that most people are only using a subset of the spec.

## Install

```
go get -u github.com/grafana/json-schema-docs
```

## Run

```
json-schema-docs -schema ./user.schema.json > user.md
```

To use a template when generating the Markdown:

```
json-schema-docs -schema ./user.schema.json -template user.md.tpl
```

`template` is the path to a file containing a Go template, such as this one:

```
+++
title = "{{ .Title }}"
description = "{{ .Description }}"
+++

# API reference

This is the reference documentation for an API.

{{ .Markdown 2 }}
```

The argument to `.Markdown` is the heading level you want the docs to start at.

## License

[Apache 2.0 License](LICENSE)
