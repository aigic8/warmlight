package models

import (
	"errors"
	"strings"
)

const CALLBACK_MSG_OUTPUTS_LIST = "olm"

const CALLBACK_MSG_NEXT_SOURCE_PAGE = "nsp"
const CALLBACK_MSG_PREV_SOURCE_PAGE = "psp"

const CALLBACK_COMMAND_ACTIVATE_OUTPUT = "ac_op"
const CALLBACK_COMMAND_DEACTIVATE_OUTPUT = "de_op"

const CALLBACK_COMMAND_SOURCE_INFO = "in_sr"
const CALLBACK_COMMAND_SOURCE_EDIT = "ed_sr"

const CALLBACK_COMMAND_MERGE_LIBRARY = "mr_lb"
const CALLBACK_COMMAND_DELETE_LIBRARY = "dl_lb"

var ErrMalformedCallbackString = errors.New("malformed callback string")

var ErrMultipleSourceKindFilters = errors.New("multiple source kinds")

type CallbackData struct {
	ReplaceMessageWith string
	Action             string
	Data               string
}

func (cd *CallbackData) Marshal() string {
	return cd.ReplaceMessageWith + "-" + cd.Action + "-" + cd.Data
}

func UnmarshalCallbackData(raw string) (CallbackData, error) {
	parts := strings.SplitN(raw, "-", 3)
	if len(parts) < 3 {
		return CallbackData{}, ErrMalformedCallbackString
	}

	return CallbackData{
		ReplaceMessageWith: parts[0],
		Action:             parts[1],
		Data:               parts[2],
	}, nil
}

type SourceFilter struct {
	Text       string
	SourceKind string
}

func ParseSourceFilter(text string) (SourceFilter, error) {
	var sf SourceFilter
	sourceKindFilters := map[string]bool{
		"@article": true,
		"@book":    true,
		"@person":  true,
		"@unknown": true,
	}

	fields := strings.Fields(text)
	sourceKindFilterIndex := -1
	for i, word := range fields {
		if _, isSourceKindFilter := sourceKindFilters[word]; isSourceKindFilter {
			if sf.SourceKind != "" {
				return sf, ErrMultipleSourceKindFilters
			}
			sf.SourceKind = strings.TrimPrefix(word, "@")
			sourceKindFilterIndex = i
		}
	}

	// TODO: use more efficient way using strings.Builder
	if sourceKindFilterIndex != 0 {
		sf.Text = fields[0]
	}

	if len(fields) > 1 {
		for i, item := range fields[1:] {
			if i+1 != sourceKindFilterIndex {
				sf.Text += " " + item
			}
		}
	}

	return sf, nil
}
