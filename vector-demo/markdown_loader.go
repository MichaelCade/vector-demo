package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Recursively load all markdown files under the given path.
func loadMarkdownFiles(rootPath string) ([]string, error) {
	var docs []string

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process markdown files
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			docs = append(docs, string(content))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

// Chunk document into smaller pieces (as before)
func chunkDocument(doc string, chunkSize int) []string {
	words := strings.Fields(doc)
	var chunks []string
	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}
	return chunks
}
