package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type invokeOutputUppercase struct {
	Text string `json:"text"`
}

var invokeRunner FunctionRunner = NewEngineRunner()

func InvokeFunctionHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		writeInvokeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method must be POST", nil, "")
		return
	}

	defer r.Body.Close()

	var req InvokeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeInvokeError(w, http.StatusBadRequest, "invalid_json", "malformed invoke request", map[string]interface{}{"error": err.Error()}, "")
		return
	}

	if strings.TrimSpace(req.Function) == "" {
		writeInvokeError(w, http.StatusUnprocessableEntity, "invalid_function", "function is required", nil, req.Meta.RequestID)
		return
	}

	spec, ok := LookupFunctionSpec(req.Function)
	if !ok {
		writeInvokeError(w, http.StatusNotFound, "unknown_function", "function is not registered", map[string]interface{}{"function": req.Function}, req.Meta.RequestID)
		return
	}

	output, invokeErr, status := executeInvoke(req)
	if invokeErr != nil {
		writeInvokeError(w, status, invokeErr.Code, invokeErr.Message, invokeErr.Details, req.Meta.RequestID)
		return
	}

	resp := InvokeResponse{
		OK:     true,
		Output: output,
		Meta: ResponseMeta{
			RequestID:  req.Meta.RequestID,
			Function:   req.Function,
			Container:  spec.Image,
			DurationMS: time.Since(start).Milliseconds(),
		},
	}

	writeInvokeResponse(w, http.StatusOK, spec.OutputMimeType, resp)
}

func executeInvoke(req InvokeRequest) (json.RawMessage, *InvokeError, int) {
	switch req.Function {
	case FunctionTextUppercase:
		var in UppercaseInput
		if err := json.Unmarshal(req.Input, &in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: "invalid input for text.uppercase", Details: map[string]interface{}{"error": err.Error()}}, http.StatusUnprocessableEntity
		}
		if err := ValidateUppercaseInput(in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: err.Error()}, http.StatusUnprocessableEntity
		}
		outBytes, _, err := invokeRunner.Run(context.Background(), FunctionRegistry[FunctionTextUppercase], []byte(in.Text))
		if err != nil {
			return nil, &InvokeError{Code: "container_execution_failed", Message: "text.uppercase failed", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		out, err := json.Marshal(invokeOutputUppercase{Text: strings.TrimSpace(string(outBytes))})
		if err != nil {
			return nil, &InvokeError{Code: "marshal_failed", Message: "failed to marshal uppercase output", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		return out, nil, http.StatusOK

	case FunctionRenderURLToPDF:
		var in URLToPDFInput
		if err := json.Unmarshal(req.Input, &in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: "invalid input for render.url_to_pdf", Details: map[string]interface{}{"error": err.Error()}}, http.StatusUnprocessableEntity
		}
		if err := ValidateURLToPDFInput(in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: err.Error()}, http.StatusUnprocessableEntity
		}
		stdin, err := json.Marshal(in)
		if err != nil {
			return nil, &InvokeError{Code: "marshal_failed", Message: "failed to marshal render input", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		stdout, _, err := invokeRunner.Run(context.Background(), FunctionRegistry[FunctionRenderURLToPDF], stdin)
		if err != nil {
			return nil, &InvokeError{Code: "container_execution_failed", Message: "render.url_to_pdf failed", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		out, err := json.Marshal(map[string]string{"data_base64": base64.StdEncoding.EncodeToString(stdout), "mime_type": "application/pdf"})
		if err != nil {
			return nil, &InvokeError{Code: "marshal_failed", Message: "failed to marshal render output", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		return out, nil, http.StatusOK

	case FunctionConvertMarkdown:
		var in MarkdownInput
		if err := json.Unmarshal(req.Input, &in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: "invalid input for convert.markdown", Details: map[string]interface{}{"error": err.Error()}}, http.StatusUnprocessableEntity
		}
		if err := ValidateMarkdownInput(in); err != nil {
			return nil, &InvokeError{Code: "invalid_input", Message: err.Error()}, http.StatusUnprocessableEntity
		}
		stdin, err := json.Marshal(in)
		if err != nil {
			return nil, &InvokeError{Code: "marshal_failed", Message: "failed to marshal markdown input", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		stdout, _, err := invokeRunner.Run(context.Background(), FunctionRegistry[FunctionConvertMarkdown], stdin)
		if err != nil {
			return nil, &InvokeError{Code: "container_execution_failed", Message: "convert.markdown failed", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		out, err := json.Marshal(map[string]string{"markdown": string(stdout)})
		if err != nil {
			return nil, &InvokeError{Code: "marshal_failed", Message: "failed to marshal markdown output", Details: map[string]interface{}{"error": err.Error()}}, http.StatusInternalServerError
		}
		return out, nil, http.StatusOK
	}

	return nil, &InvokeError{Code: "unknown_function", Message: fmt.Sprintf("unknown function %q", req.Function)}, http.StatusNotFound
}

func writeInvokeError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}, requestID string) {
	resp := InvokeResponse{
		OK:    false,
		Error: &InvokeError{Code: code, Message: message, Details: details},
		Meta:  ResponseMeta{RequestID: requestID},
	}
	writeInvokeResponse(w, status, "application/json; charset=UTF-8", resp)
}

func writeInvokeResponse(w http.ResponseWriter, status int, outputContentType string, resp InvokeResponse) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if outputContentType != "" {
		w.Header().Set("X-Function-Output-Content-Type", outputContentType)
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
