package out

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteError writes an error message to the given writer.
// If asJSON is true, outputs JSON format; otherwise human-readable.
func WriteError(w io.Writer, asJSON bool, err error) error {
	if err == nil {
		return nil
	}

	if asJSON {
		return json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}

	_, writeErr := fmt.Fprintf(w, "Error: %v\n", err)
	return writeErr
}

// WriteJSON writes a value as JSON to the given writer.
func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
