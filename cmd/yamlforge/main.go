// Package main provides the YAML Forge CLI.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/convert"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/diff"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/filter"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/format"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/lint"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/merge"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/parser"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/query"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/stats"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/validate"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "yamlforge",
		Short: "YAML Forge - Comprehensive YAML Processing Toolkit",
		Long:  "A powerful CLI tool for parsing, querying, diffing, merging, converting, linting, formatting, and validating YAML documents.",
	}

	// Format command
	formatCmd := &cobra.Command{
		Use:   "fmt [file]",
		Short: "Format YAML files",
		Long:  "Format and pretty-print YAML files with consistent indentation.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			indent, _ := cmd.Flags().GetInt("indent")
			sortKeys, _ := cmd.Flags().GetBool("sort-keys")
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			f := format.NewFormatter(format.Options{Indent: indent, SortKeys: sortKeys})
			result, err := f.Format(data)
			if err != nil {
				return err
			}
			fmt.Print(string(result))
			return nil
		},
	}
	formatCmd.Flags().IntP("indent", "i", 2, "indentation spaces")
	formatCmd.Flags().BoolP("sort-keys", "s", false, "sort keys alphabetically")
	rootCmd.AddCommand(formatCmd)

	// Lint command
	lintCmd := &cobra.Command{
		Use:   "lint [file]",
		Short: "Lint YAML files",
		Long:  "Check YAML files for common issues and best practices.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := lint.NewEngine()
			report, err := engine.LintFile(args[0])
			if err != nil {
				return err
			}

			if len(report.Issues) == 0 {
				fmt.Println("✓ No issues found")
				return nil
			}

			for _, issue := range report.Issues {
				icon := "ℹ"
				switch issue.Severity {
				case "error":
					icon = "✗"
				case "warning":
					icon = "⚠"
				}
				fmt.Printf("%s [%d] %s: %s\n", icon, issue.Line, issue.Rule, issue.Message)
			}

			fmt.Printf("\n%d errors, %d warnings\n", report.ErrorCount(), report.WarningCount())
			if report.HasErrors() {
				os.Exit(1)
			}
			return nil
		},
	}
	rootCmd.AddCommand(lintCmd)

	// Query command
	queryCmd := &cobra.Command{
		Use:   "query [file] [path]",
		Short: "Query YAML with path expressions",
		Long:  "Query YAML documents using dot-notation path expressions.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := parser.New()
			docs, err := p.ParseFile(args[0])
			if err != nil {
				return err
			}
			if len(docs) == 0 {
				return fmt.Errorf("no documents found")
			}

			engine := query.NewEngine()
			results, err := engine.Query(docs[0], args[1])
			if err != nil {
				return err
			}

			for _, node := range results {
				fmt.Printf("%v\n", node.Value)
			}
			return nil
		},
	}
	rootCmd.AddCommand(queryCmd)

	// Diff command
	diffCmd := &cobra.Command{
		Use:   "diff [file1] [file2]",
		Short: "Diff two YAML files",
		Long:  "Compare two YAML files and show structural differences.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			p := parser.New()

			docs1, err := p.ParseFile(args[0])
			if err != nil {
				return err
			}
			docs2, err := p.ParseFile(args[1])
			if err != nil {
				return err
			}

			var node1, node2 *models.Node
			if len(docs1) > 0 {
				node1 = docs1[0]
			}
			if len(docs2) > 0 {
				node2 = docs2[0]
			}

			engine := diff.NewEngine()
			entries := engine.Diff(node1, node2)

			if !diff.HasChanges(entries) {
				fmt.Println("Files are identical")
				return nil
			}

			added, removed, modified := diff.Summary(entries)
			fmt.Print(diff.FormatDiff(entries, format))
			fmt.Printf("\n%d added, %d removed, %d modified\n", added, removed, modified)
			return nil
		},
	}
	diffCmd.Flags().StringP("format", "f", "text", "output format (text, json, compact)")
	rootCmd.AddCommand(diffCmd)

	// Merge command
	mergeCmd := &cobra.Command{
		Use:   "merge [file1] [file2]",
		Short: "Merge two YAML files",
		Long:  "Deep merge two YAML files with configurable strategy.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			strategy, _ := cmd.Flags().GetString("strategy")
			p := parser.New()

			docs1, err := p.ParseFile(args[0])
			if err != nil {
				return err
			}
			docs2, err := p.ParseFile(args[1])
			if err != nil {
				return err
			}

			var node1, node2 *models.Node
			if len(docs1) > 0 {
				node1 = docs1[0]
			}
			if len(docs2) > 0 {
				node2 = docs2[0]
			}

			var mergeStrategy models.MergeStrategy
			switch strategy {
			case "replace":
				mergeStrategy = models.MergeReplace
			case "append":
				mergeStrategy = models.MergeAppend
			case "keep":
				mergeStrategy = models.MergeKeep
			default:
				mergeStrategy = models.MergeDeep
			}

			engine := merge.NewEngine(mergeStrategy)
			result, err := engine.Merge(node1, node2)
			if err != nil {
				return err
			}

			f := format.NewFormatter(format.DefaultOptions())
			data, err := f.FormatNode(result)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}
	mergeCmd.Flags().StringP("strategy", "s", "deep", "merge strategy (deep, replace, append, keep)")
	rootCmd.AddCommand(mergeCmd)

	// Convert command
	convertCmd := &cobra.Command{
		Use:   "convert [file]",
		Short: "Convert YAML to other formats",
		Long:  "Convert YAML files to JSON, TOML, or CSV.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			to, _ := cmd.Flags().GetString("to")
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			engine := convert.NewEngine()
			var result []byte

			switch to {
			case "json":
				result, err = engine.ToJSON(data)
			case "json-compact":
				result, err = engine.ToJSONCompact(data)
			case "toml":
				result, err = engine.ToTOML(data)
			case "csv":
				result, err = engine.ToCSV(data)
			default:
				return fmt.Errorf("unsupported format: %s (use json, toml, csv)", to)
			}

			if err != nil {
				return err
			}
			fmt.Print(string(result))
			return nil
		},
	}
	convertCmd.Flags().StringP("to", "t", "json", "target format (json, toml, csv, json-compact)")
	rootCmd.AddCommand(convertCmd)

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate [file] [schema]",
		Short: "Validate YAML against a schema",
		Long:  "Validate YAML files against a schema definition.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			schema, err := validate.LoadSchema(args[1])
			if err != nil {
				return err
			}

			engine := validate.NewEngine()
			report, err := engine.ValidateFile(args[0], schema)
			if err != nil {
				return err
			}

			if len(report.Issues) == 0 {
				fmt.Println("✓ Validation passed")
				return nil
			}

			for _, issue := range report.Issues {
				icon := "✗"
				if issue.Severity == "warning" {
					icon = "⚠"
				}
				fmt.Printf("%s [%d] %s: %s\n", icon, issue.Line, issue.Rule, issue.Message)
			}

			fmt.Printf("\n%d errors, %d warnings\n", report.ErrorCount(), report.WarningCount())
			if report.HasErrors() {
				os.Exit(1)
			}
			return nil
		},
	}
	rootCmd.AddCommand(validateCmd)

	// Filter command
	filterCmd := &cobra.Command{
		Use:   "filter [file]",
		Short: "Filter multi-document YAML",
		Long:  "Filter YAML documents from multi-document files.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filterType, _ := cmd.Flags().GetString("type")
			value, _ := cmd.Flags().GetString("value")

			engine := filter.NewEngine()
			result, err := engine.FilterFile(args[0], filterType, value)
			if err != nil {
				return err
			}
			fmt.Print(string(result))
			return nil
		},
	}
	filterCmd.Flags().StringP("type", "t", "content", "filter type (path, content, index)")
	filterCmd.Flags().StringP("value", "v", "", "filter value")
	rootCmd.AddCommand(filterCmd)

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats [file]",
		Short: "Analyze YAML structure",
		Long:  "Compute statistics about YAML document structure.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")

			engine := stats.NewEngine()
			s, err := engine.AnalyzeFile(args[0])
			if err != nil {
				return err
			}

			switch output {
			case "json":
				data, err := json.MarshalIndent(stats.StatsToMap(s), "", "  ")
				if err != nil {
					return err
				}
				fmt.Print(string(data))
			default:
				fmt.Print(stats.StatsString(s))
			}
			return nil
		},
	}
	statsCmd.Flags().StringP("output", "o", "text", "output format (text, json)")
	rootCmd.AddCommand(statsCmd)

	// Infer schema command
	schemaCmd := &cobra.Command{
		Use:   "schema [file]",
		Short: "Infer schema from YAML",
		Long:  "Infer a JSON Schema from a YAML file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			schema, err := validate.InferSchema(data)
			if err != nil {
				return err
			}

			schemaJSON, err := json.MarshalIndent(schema, "", "  ")
			if err != nil {
				return err
			}
			fmt.Print(string(schemaJSON))
			return nil
		},
	}
	rootCmd.AddCommand(schemaCmd)

	// Count command
	countCmd := &cobra.Command{
		Use:   "count [file]",
		Short: "Count YAML documents",
		Long:  "Count the number of YAML documents in a file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			engine := filter.NewEngine()
			count, err := engine.CountDocuments(data)
			if err != nil {
				return err
			}
			fmt.Printf("%d document(s)\n", count)
			return nil
		},
	}
	rootCmd.AddCommand(countCmd)

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("yamlforge %s\n", version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Sort command
	sortCmd := &cobra.Command{
		Use:   "sort [file]",
		Short: "Sort YAML keys alphabetically",
		Long:  "Sort all mapping keys in a YAML file alphabetically.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			result, err := format.SortKeys(data)
			if err != nil {
				return err
			}
			fmt.Print(string(result))
			return nil
		},
	}
	rootCmd.AddCommand(sortCmd)

	// Print command (pretty-print)
	printCmd := &cobra.Command{
		Use:   "print [file]",
		Short: "Pretty-print YAML",
		Long:  "Pretty-print YAML with proper formatting.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			result, err := format.PrettyPrint(data, 2)
			if err != nil {
				return err
			}
			fmt.Print(string(result))
			return nil
		},
	}
	rootCmd.AddCommand(printCmd)

	// Keys command
	keysCmd := &cobra.Command{
		Use:   "keys [file]",
		Short: "List all keys in YAML",
		Long:  "List all keys in a YAML file with their paths.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			s, err := stats.NewEngine().Analyze(data)
			if err != nil {
				return err
			}

			// Get unique keys
			seen := make(map[string]bool)
			for key := range s.KeyFrequency {
				if !seen[key] {
					seen[key] = true
					count := s.KeyFrequency[key]
					fmt.Printf("%s (used %d time(s))\n", key, count)
				}
			}
			return nil
		},
	}
	rootCmd.AddCommand(keysCmd)

	// Flatten command
	flattenCmd := &cobra.Command{
		Use:   "flatten [file]",
		Short: "Flatten nested YAML to dot-notation",
		Long:  "Convert nested YAML to flat dot-notation keys.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var doc map[string]interface{}
			if err := json.Unmarshal(data, &doc); err != nil {
				// Try YAML
				if err2 := yaml.Unmarshal(data, &doc); err2 != nil {
					return fmt.Errorf("parsing YAML: %w", err2)
				}
			}

			flattened := flattenMap(doc, "")
			for key, value := range flattened {
				fmt.Printf("%s: %v\n", key, value)
			}
			return nil
		},
	}
	rootCmd.AddCommand(flattenCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// flattenMap flattens a nested map to dot-notation.
func flattenMap(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			for k, val := range flattenMap(v, fullKey) {
				result[k] = val
			}
		case []interface{}:
			for i, item := range v {
				itemKey := fmt.Sprintf("%s[%d]", fullKey, i)
				if subMap, ok := item.(map[string]interface{}); ok {
					for k, val := range flattenMap(subMap, itemKey) {
						result[k] = val
					}
				} else {
					result[itemKey] = item
				}
			}
		default:
			result[fullKey] = value
		}
	}
	return result
}
