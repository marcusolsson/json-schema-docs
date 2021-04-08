package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestSchema(t *testing.T) {
	schemaTests := []struct {
		name   string
		schema string
		level  int
	}{
		{name: "address", schema: "address.schema.json"},
		{name: "arrays", schema: "arrays.schema.json"},
		{name: "basic", schema: "basic.schema.json"},
		{name: "calendar", schema: "calendar.schema.json"},
		{name: "card", schema: "card.schema.json"},
		{name: "geographical-location", schema: "geographical-location.schema.json"},
		{name: "ref-hell", schema: "ref-hell.schema.json"},
		{name: "deep-headings", schema: "ref-hell.schema.json", level: 5},
	}

	for _, tt := range schemaTests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.schema))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			schema, err := newSchema(f, "testdata")
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			if tt.level > 0 {
				buf.WriteString(schema.Markdown(tt.level))
			} else {
				buf.WriteString(schema.Markdown(1))
			}

			gp := filepath.Join("testdata", strings.Replace(t.Name()+".golden", "/", "_", -1))
			if *update {
				if err := ioutil.WriteFile(gp, buf.Bytes(), 0644); err != nil {
					t.Fatal("failed to update golden ")
				}
			}

			g, err := ioutil.ReadFile(gp)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(buf.Bytes(), g) {
				t.Log(buf.String())
				t.Errorf("data does not match .golden file")
			}
		})
	}

}
