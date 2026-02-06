package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/spf13/cobra"
)

type apiOptions struct {
	method   string
	hostname string
	headers  []string
	field    []string
	rawField []string
	input    string
	silent   bool
	jq       string
}

func newAPICmd() *cobra.Command {
	opts := &apiOptions{}

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated API request",
		Long: `Makes an authenticated HTTP request to the GitFlic API and prints the response.

The endpoint argument should be a path like "/user/me" or "/project/owner/repo".
The default HTTP request method is GET, use --method to change it.`,
		Example: `  # Get current user
  gf api /user/me

  # List projects
  gf api /project

  # Create an issue
  gf api /project/owner/repo/issue --method POST -f title="Bug report" -f description="Details"

  # Get with raw JSON body
  gf api /project/owner/repo/issue --method POST --input body.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPI(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.method, "method", "X", "GET", "HTTP method")
	cmd.Flags().StringVarP(&opts.hostname, "hostname", "H", "", "GitFlic hostname")
	cmd.Flags().StringArrayVar(&opts.headers, "header", nil, "Add HTTP headers (key:value)")
	cmd.Flags().StringArrayVarP(&opts.field, "field", "f", nil, "Add JSON field (key=value)")
	cmd.Flags().StringArrayVarP(&opts.rawField, "raw-field", "F", nil, "Add raw JSON field (key=value, value is raw JSON)")
	cmd.Flags().StringVar(&opts.input, "input", "", "Read request body from file")
	cmd.Flags().BoolVar(&opts.silent, "silent", false, "Do not print response body")
	cmd.Flags().StringVarP(&opts.jq, "jq", "q", "", "Filter response with jq expression (simple: .field, .field.subfield)")

	return cmd
}

// validHTTPMethods contains allowed HTTP methods
var validHTTPMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"PATCH":   true,
	"DELETE":  true,
	"HEAD":    true,
	"OPTIONS": true,
}

func runAPI(opts *apiOptions, endpoint string) error {
	// Validate HTTP method
	method := strings.ToUpper(opts.method)
	if !validHTTPMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s (allowed: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)", opts.method)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	token, err := cfg.Token()
	if err != nil {
		return fmt.Errorf("not authenticated. Run 'gf auth login' first")
	}

	hostname := opts.hostname
	if hostname == "" {
		hostname = cfg.ActiveHost
	}
	if hostname == "" {
		hostname = config.DefaultHost()
	}

	client := api.NewClient(config.BaseURL(hostname), token)

	// Build request body
	var body any
	if opts.input != "" {
		// Read from file
		data, err := os.ReadFile(opts.input)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		var jsonBody any
		if err := json.Unmarshal(data, &jsonBody); err != nil {
			return fmt.Errorf("invalid JSON in input file: %w", err)
		}
		body = jsonBody
	} else if len(opts.field) > 0 || len(opts.rawField) > 0 {
		// Build JSON from fields
		bodyMap := make(map[string]any)

		for _, f := range opts.field {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid field format: %s (use key=value)", f)
			}
			bodyMap[parts[0]] = parts[1]
		}

		for _, f := range opts.rawField {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid raw-field format: %s (use key=value)", f)
			}
			var value any
			if err := json.Unmarshal([]byte(parts[1]), &value); err != nil {
				return fmt.Errorf("invalid JSON in raw-field %s: %w", parts[0], err)
			}
			bodyMap[parts[0]] = value
		}

		body = bodyMap
	}

	// Make request
	var response json.RawMessage
	err = client.REST(method, endpoint, body, &response)
	if err != nil {
		return err
	}

	if opts.silent {
		return nil
	}

	// Handle jq filter
	if opts.jq != "" {
		response, err = simpleJQ(response, opts.jq)
		if err != nil {
			return fmt.Errorf("jq filter error: %w", err)
		}
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, response, "", "  "); err != nil {
		// If not valid JSON, just print as-is
		fmt.Println(string(response))
	} else {
		io.Copy(os.Stdout, &prettyJSON)
		fmt.Println()
	}

	return nil
}

// simpleJQ implements a very basic jq-like filter
// Supports: .field, .field.subfield, .[0], .field[0]
func simpleJQ(data json.RawMessage, filter string) (json.RawMessage, error) {
	if filter == "." {
		return data, nil
	}

	filter = strings.TrimPrefix(filter, ".")

	var current any
	if err := json.Unmarshal(data, &current); err != nil {
		return nil, err
	}

	parts := strings.Split(filter, ".")
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Check for array index
		if strings.Contains(part, "[") {
			bracketIdx := strings.Index(part, "[")
			fieldName := part[:bracketIdx]
			indexStr := strings.Trim(part[bracketIdx:], "[]")

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", indexStr)
			}

			if fieldName != "" {
				m, ok := current.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("cannot access field %s on non-object", fieldName)
				}
				current = m[fieldName]
			}

			arr, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("cannot index non-array")
			}
			if index < 0 || index >= len(arr) {
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			}
			current = arr[index]
		} else {
			m, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("cannot access field %s on non-object", part)
			}
			current = m[part]
		}
	}

	return json.Marshal(current)
}
