package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/olekukonko/tablewriter"
)

var errCrossSchemaReference = errors.New("cross-schema reference")

type schema struct {
	ID          string             `json:"$id,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Schema      string             `json:"$schema,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Type        string             `json:"type,omitempty"`
	Properties  map[string]*schema `json:"properties,omitempty"`
	Items       *schema            `json:"items,omitempty"`
	Definitions map[string]*schema `json:"definitions,omitempty"`
}

func newSchema(r io.Reader, workingDir string) (*schema, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var data schema
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	// Needed for resolving in-schema references.
	root, err := simplejson.NewJson(b)
	if err != nil {
		return nil, err
	}

	return resolveSchema(&data, workingDir, root)
}

// Markdown returns the Markdown representation of the schema.
//
// The level argument can be used to offset the heading levels. This can be
// useful if you want to add the schema under a subheading.
func (s schema) Markdown(level int) string {
	if level < 1 {
		level = 1
	}

	var buf bytes.Buffer

	if s.Title != "" {
		fmt.Fprintln(&buf, strings.Repeat("#", level)+" "+s.Title)
		fmt.Fprintln(&buf)
	}

	if s.Description != "" {
		fmt.Fprintln(&buf, s.Description)
		fmt.Fprintln(&buf)
	}

	if len(s.Properties) > 0 {
		fmt.Fprintln(&buf, strings.Repeat("#", level+1)+" Properties")
		fmt.Fprintln(&buf)
	}

	printProperties(&buf, &s)

	// Add padding.
	fmt.Fprintln(&buf)

	for _, obj := range findDefinitions(&s) {
		fmt.Fprintf(&buf, obj.Markdown(level+1))
	}

	return buf.String()
}

func findDefinitions(s *schema) []*schema {
	// Gather all properties of object type so that we can generate the
	// properties for them recursively.
	var objs []*schema

	for k, p := range s.Properties {
		// Use the identifier as the title.
		if p.Type == "object" {
			p.Title = k
			objs = append(objs, p)
		}

		// If the property is an array of objects, use the name of the array
		// property as the title.
		if p.Type == "array" {
			if p.Items != nil {
				if p.Items.Type == "object" {
					p.Items.Title = k
					objs = append(objs, p.Items)
				}
			}
		}
	}

	// Sort the object schemas.
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].Title < objs[j].Title
	})

	return objs
}

func printProperties(w io.Writer, s *schema) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Property", "Type", "Required", "Description"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	// Buffer all property rows so that we can sort them before printing them.
	var rows [][]string

	for k, p := range s.Properties {
		// Generate relative links for objects and arrays of objects.
		var propType string
		switch p.Type {
		case "object":
			propType = fmt.Sprintf("[%s](#%s)", p.Type, strings.ToLower(k))
		case "array":
			if p.Items != nil {
				if p.Items.Type == "object" {
					propType = fmt.Sprintf("[%s](#%s)[]", p.Items.Type, strings.ToLower(k))
				} else {
					propType = fmt.Sprintf("%s[]", p.Items.Type)
				}
			} else {
				propType = p.Type
			}
		default:
			propType = p.Type
		}

		// Emphasize required properties.
		var required string
		if in(s.Required, k) {
			required = "**Yes**"
		} else {
			required = "No"
		}

		rows = append(rows, []string{fmt.Sprintf("`%s`", k), propType, required, p.Description})
	}

	// Sort by the required column, then by the name column.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][2] < rows[j][2] {
			return true
		}
		if rows[i][2] > rows[j][2] {
			return false
		}
		return rows[i][0] < rows[j][0]
	})

	table.AppendBulk(rows)
	table.Render()
}

// in returns true if a string slice contains a specific string.
func in(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
