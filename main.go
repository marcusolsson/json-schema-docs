package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

func main() {
	var (
		schemaPath   = flag.String("schema", "", "Path to the JSON Schema")
		templatePath = flag.String("template", "", "Path to a template")
	)

	flag.Parse()

	if *schemaPath == "" {
		fmt.Println("no path to schema")
		os.Exit(1)
	}

	f, err := os.Open(*schemaPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	schema, err := newSchema(f)
	if err != nil {
		log.Fatal(err)
	}

	tpl, err := getOrDefaultTemplate(*templatePath)
	if err != nil {
		log.Fatal(err)
	}

	if err := tpl.Execute(os.Stdout, schema); err != nil {
		log.Fatal(err)
	}
}

func getOrDefaultTemplate(path string) (*template.Template, error) {
	if path == "" {
		return template.New("docs").Parse(`{{ .Markdown 1 }}`)
	}
	return template.ParseFiles(path)
}
