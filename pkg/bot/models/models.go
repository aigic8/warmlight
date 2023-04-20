package models

const OUTPUTS_LIST_MSG = "OLM"

type CallbackData struct {
	DeactivateOutputs  []int64 `json:"deactivateOutputs"`
	ActivateOutputs    []int64 `json:"activateOutputs"`
	ReplaceMessageWith string  `json:"replaceMessageWith"`
}
