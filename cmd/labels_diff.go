package cmd

// applyLabelsDiff applies gcloud's labels Diff (clear, then remove, then
// update) to current and reports whether the result differs.
func applyLabelsDiff(current map[string]string, clear bool, remove []string, update map[string]string) (map[string]string, bool) {
	next := map[string]string{}
	if !clear {
		for k, v := range current {
			next[k] = v
		}
	}
	for _, k := range remove {
		delete(next, k)
	}
	for k, v := range update {
		next[k] = v
	}
	if len(next) != len(current) {
		return next, true
	}
	for k, v := range next {
		if cv, ok := current[k]; !ok || cv != v {
			return next, true
		}
	}
	return next, false
}
