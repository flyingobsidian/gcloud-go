package config

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// iniFile is a gcloud configuration file held as its original lines, so that
// setting or unsetting one property leaves every other section, key and
// comment as it was. Python gcloud gets the same preservation of sections and
// keys from configparser (properties_file.PersistProperty).
type iniFile struct {
	lines []string
}

// readINIFile reads the INI file at path. A missing file yields an empty
// iniFile.
func readINIFile(path string) (*iniFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &iniFile{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	s := strings.TrimSuffix(string(data), "\n")
	if s == "" {
		return &iniFile{}, nil
	}
	return &iniFile{lines: strings.Split(s, "\n")}, nil
}

// write saves the file to path.
func (f *iniFile) write(path string) error {
	var s string
	if len(f.lines) > 0 {
		s = strings.Join(f.lines, "\n") + "\n"
	}
	return os.WriteFile(path, []byte(s), 0600)
}

// parseINILine classifies line as a section header (isSection, name set), a
// key/value pair (isKey, key and val set) or neither (blank, comment or
// unparseable).
func parseINILine(line string) (name, key, val string, isSection, isKey bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return "", "", "", false, false
	}
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return line[1 : len(line)-1], "", "", true, false
	}
	k, v, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", "", false, false
	}
	return "", strings.TrimSpace(k), strings.TrimSpace(v), false, true
}

// get returns the value of key in section, or "" when absent. If the key is
// repeated, the last occurrence wins.
func (f *iniFile) get(section, key string) string {
	val := ""
	f.each(func(s, k, v string) {
		if s == section && k == key {
			val = v
		}
	})
	return val
}

// each calls fn for every key/value pair in file order.
func (f *iniFile) each(fn func(section, key, val string)) {
	current := ""
	for _, line := range f.lines {
		name, k, v, isSection, isKey := parseINILine(line)
		switch {
		case isSection:
			current = name
		case isKey:
			fn(current, k, v)
		}
	}
}

// set sets key in section to value, replacing the existing line in place,
// appending the key to the end of the section, or appending a new section to
// the end of the file.
func (f *iniFile) set(section, key, value string) {
	entry := key + " = " + value
	current := ""
	sectionFound := false
	insertAt := 0 // index after the header or last key of section
	for i, line := range f.lines {
		name, k, _, isSection, isKey := parseINILine(line)
		if isSection {
			current = name
			if current == section && !sectionFound {
				sectionFound = true
				insertAt = i + 1
			}
			continue
		}
		if current != section || !isKey {
			continue
		}
		if k == key {
			f.lines[i] = entry
			f.removeKey(section, key, i+1)
			return
		}
		insertAt = i + 1
	}
	if sectionFound {
		f.lines = slices.Insert(f.lines, insertAt, entry)
		return
	}
	if n := len(f.lines); n > 0 && strings.TrimSpace(f.lines[n-1]) != "" {
		f.lines = append(f.lines, "")
	}
	f.lines = append(f.lines, "["+section+"]", entry)
}

// unset removes key from section. The section header is kept, as Python
// gcloud does.
func (f *iniFile) unset(section, key string) {
	f.removeKey(section, key, 0)
}

// removeKey deletes every line for key in section at or after index from.
func (f *iniFile) removeKey(section, key string, from int) {
	current := ""
	kept := f.lines[:0:0]
	for i, line := range f.lines {
		name, k, _, isSection, isKey := parseINILine(line)
		if isSection {
			current = name
		}
		if i >= from && isKey && current == section && k == key {
			continue
		}
		kept = append(kept, line)
	}
	f.lines = kept
}
