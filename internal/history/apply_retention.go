package history

import (
	"fmt"
	"io"
	"os"
)

// ApplyRetention applies the given RetentionPolicy to all checks stored
// under the store's directory. It delegates to Prune with the policy values.
// Progress and errors are written to w (pass io.Discard to silence output).
func ApplyRetention(s *Store, policy RetentionPolicy, w io.Writer) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("apply retention: %w", err)
	}

	if policy.IsZero() {
		fmt.Fprintln(w, "retention: no constraints set, skipping prune")
		return nil
	}

	opts := PruneOptions{
		Dir:        s.Dir(),
		MaxEntries: policy.MaxEntries,
		MaxAge:     policy.MaxAge,
	}

	removed, err := Prune(opts)
	if err != nil {
		return fmt.Errorf("apply retention: %w", err)
	}

	if removed == 0 {
		fmt.Fprintln(w, "retention: nothing to prune")
	} else {
		fmt.Fprintf(w, "retention: pruned %d entr%s\n", removed, pluralSuffix(removed))
	}
	return nil
}

func pluralSuffix(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

// Dir exposes the underlying directory of the Store for use by other packages.
func (s *Store) Dir() string {
	return s.dir
}

// PruneOptions holds parameters for a Prune call (re-exported for clarity).
type PruneOptions = pruneOptions

// Ensure the Store has the dir field accessible. This file adds the Dir()
// accessor; the field itself lives in store.go.
var _ = os.DevNull
