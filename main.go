package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/nick-jones/gost/pkg/scan"
)

const tmpl = `{{printf "%x: %q" .Addr .Value}} → {{range $i, $e := .Refs}}
{{- if le $i 5}}{{ printf "%s:%d " .File .Line }}{{end}}
{{- end}}
{{- if gt (len .Refs) 5}}... (truncated, {{len .Refs}} total){{- end -}}
`

func main() {
	app := &cli.App{
		Name: "gost",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "template",
				Usage: "template string for printing the results (format is text/template)",
				Value: tmpl,
			},
			&cli.StringFlag{
				Name:  "string-table",
				Usage: `if symbols are missing, use values "guess" or "ignore" to enable more fuzzy matching`,
			},
			&cli.BoolFlag{
				Name:  "no-nulls",
				Usage: "strings containing null characters will be ignored",
				Value: true,
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	filePath := c.Args().First()
	format := c.String("template")

	tmpl, err := template.New("format").Parse(format)
	if err != nil {
		return fmt.Errorf("failed to parse format flag: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	opts, err := parseFlags(c)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// run analysis
	results, err := scan.Run(f, opts...)
	if err != nil {
		return fmt.Errorf("failed to search instructions: %w", err)
	}

	// print results
	for _, res := range results {
		if err := tmpl.Execute(os.Stdout, res); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		fmt.Println()
	}

	return nil
}

func parseFlags(c *cli.Context) ([]scan.Option, error) {
	opts := make([]scan.Option, 0)

	switch flag := c.String("string-table"); flag {
	case "guess":
		opts = append(opts, scan.WithStringTableGuessed())
	case "ignore":
		opts = append(opts, scan.WithStringTableIgnored())
	case "":
	default:
		return nil, fmt.Errorf("invalid str-table flag value: %s", flag)
	}

	if c.Bool("no-nulls") {
		opts = append(opts, scan.WithNoNulls())
	}

	return opts, nil
}
