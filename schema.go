package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type schema struct {
	ID          string            `json:"$id"`
	Schema      string            `json:"$schema"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Required    []string          `json:"required"`
	Type        string            `json:"type"`
	Properties  map[string]schema `json:"properties"`
	Items       *schema           `json:"items"`
}

func newSchema(r io.Reader) (*schema, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var data schema
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return &data, nil
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

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Property", "Type", "Required", "Description"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	// Gather all properties of object type so that we can generate the
	// properties for them recursively.
	var objs []schema

	// Buffer all property rows so that we can sort them before printing them.
	var rows [][]string

	for k, p := range s.Properties {
		// Use the identifier as the title.
		if p.Type == "object" {
			p.Title = k
			objs = append(objs, p)
		}

		// If the property is an array of objects, use the name of the array
		// property as the title.
		if p.Type == "array" && p.Items != nil {
			if p.Items.Type == "object" {
				p.Items.Title = k
				objs = append(objs, *p.Items)
			}
		}

		// Generate relative links for objects and arrays of objects.
		var propType string
		switch p.Type {
		case "object":
			propType = fmt.Sprintf("[%s](#%s)", p.Type, strings.ToLower(k))
		case "array":
			if p.Items != nil {
				propType = fmt.Sprintf("[%s](#%s)", p.Type, strings.ToLower(k))
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

	// Add padding.
	fmt.Fprintln(&buf)

	// Sort the object schemas before recursing.
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].Title < objs[j].Title
	})
	for _, obj := range objs {
		fmt.Fprintf(&buf, obj.Markdown(level+1))
	}

	return buf.String()
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
