package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/nick-jones/gost/pkg/analysis"
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

	// run analysis
	results, err := analysis.Run(f)
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
