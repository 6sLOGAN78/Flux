// Package utils provides general serialization helpers.
package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON writes indented JSON to the process output for local inspection.
func PrintJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = fmt.Fprintln(os.Stdout, "JSON:", string(data))
	return err
}
