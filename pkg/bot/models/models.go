package models

import (
	"errors"
	"strings"
)

const CALLBACK_OUTPUTS_LIST_MSG = "olm"

const CALLBACK_COMMAND_ACTIVATE_OUTPUT = "ac_op"
const CALLBACK_COMMAND_DEACTIVATE_OUTPUT = "de_op"

var ErrMalformedCallbackString = errors.New("malformed callback string")

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
