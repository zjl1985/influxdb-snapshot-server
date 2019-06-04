package models

type Result struct {
	Success bool   `json:"success"`
	Result  interface{} `json:"result"`
}
