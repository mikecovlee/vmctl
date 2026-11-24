//go:build ignore

package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"vmctl/internal/config"
)

var linkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^\s)]+)(?:\s+"[^"]*")?\)`)

func main() {
	if err := check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Example configurations and local Markdown links passed")
}

func check() error {
	return filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "generated" || entry.Name() == "dist") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(path, "examples/") && filepath.Ext(path) == ".json" {
			if _, err := config.Read(path); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(data), -1) {
			target, err := url.Parse(match[1])
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if target.IsAbs() || target.Host != "" {
				continue
			}
			file := path
			if target.Path != "" {
				file = filepath.Join(filepath.Dir(path), target.Path)
			}
			if _, err := os.Stat(file); err != nil {
				return fmt.Errorf("%s: broken link %s: %w", path, match[1], err)
			}
			if target.Fragment != "" && strings.HasSuffix(file, ".md") {
				contents, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				if !anchors(string(contents))[target.Fragment] {
					return fmt.Errorf("%s: missing anchor %s", path, match[1])
				}
			}
		}
		return nil
	})
}

func anchors(markdown string) map[string]bool {
	result := make(map[string]bool)
	counts := make(map[string]int)
	fenced := false
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
		}
		if fenced || !strings.HasPrefix(line, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
		slug := strings.Map(func(r rune) rune {
			switch {
			case unicode.IsSpace(r):
				return '-'
			case unicode.IsLetter(r), unicode.IsNumber(r), r == '-', r == '_':
				return unicode.ToLower(r)
			default:
				return -1
			}
		}, heading)
		anchor := slug
		if counts[slug] > 0 {
			anchor = fmt.Sprintf("%s-%d", slug, counts[slug])
		}
		counts[slug]++
		result[anchor] = true
	}
	return result
}
