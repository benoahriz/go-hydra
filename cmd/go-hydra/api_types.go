package main

import "encoding/json"

type InvokeRequest struct {
	Function string          `json:"function"`
	Input    json.RawMessage `json:"input"`
	Meta     RequestMeta     `json:"meta,omitempty"`
}

type RequestMeta struct {
	RequestID string `json:"request_id,omitempty"`
	TimeoutMS int    `json:"timeout_ms,omitempty"`
}

type InvokeResponse struct {
	OK     bool            `json:"ok"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  *InvokeError    `json:"error,omitempty"`
	Meta   ResponseMeta    `json:"meta,omitempty"`
}

type InvokeError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type ResponseMeta struct {
	RequestID  string `json:"request_id,omitempty"`
	Function   string `json:"function,omitempty"`
	Container  string `json:"container,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}

type UppercaseInput struct {
	Text string `json:"text"`
}

type URLToPDFInput struct {
	URL string `json:"url"`
}

type MarkdownInput struct {
	URL  string `json:"url,omitempty"`
	HTML string `json:"html,omitempty"`
}
