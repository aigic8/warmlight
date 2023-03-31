package utils

import (
	"errors"
	"strings"
)

type Quote struct {
	Text       string
	MainSource string
	Sources    []string
	Tags       []string
}

// TODO maybe test?
func ParseQuote(text string) (*Quote, error) {
	lines := strings.Split(text, "\n")
	q := Quote{}

	sourcesMap := map[string]bool{}
	tagsMap := map[string]bool{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if q.Text == "" {
			q.Text = line
			continue
		}

		if strings.HasPrefix(line, "sources:") {
			sources := strings.Split(strings.TrimPrefix(line, "sources:"), ",")
			for _, sourceRaw := range sources {
				source := strings.TrimSpace(sourceRaw)
				if q.MainSource == "" {
					q.MainSource = source
				}
				sourcesMap[source] = true
			}
		}

		words := strings.Fields(line)
		for _, word := range words {
			if strings.HasPrefix(word, "#") {
				tagsMap[word[1:]] = true
			}
		}
	}

	if q.Text == "" {
		return nil, errors.New("quote has no text")
	}

	for source := range sourcesMap {
		q.Sources = append(q.Sources, source)
	}

	for tag := range tagsMap {
		q.Tags = append(q.Tags, tag)
	}

	return &q, nil
}
