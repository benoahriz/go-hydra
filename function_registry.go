package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

const (
	FunctionTextUppercase = "text.uppercase"
	FunctionRenderURLToPDF = "render.url_to_pdf"
	FunctionConvertMarkdown = "convert.markdown"
)

type FunctionSpec struct {
	Name           string
	Image          string
	Command        []string
	TimeoutMS      int
	OutputMimeType string
}

type FunctionRunner interface {
	Run(ctx context.Context, spec FunctionSpec, stdin []byte) (stdout []byte, stderr []byte, err error)
}

var FunctionRegistry = map[string]FunctionSpec{
	FunctionTextUppercase: {
		Name:           FunctionTextUppercase,
		Image:          "busybox",
		Command:        []string{"tr", "[:lower:]", "[:upper:]"},
		TimeoutMS:      30000,
		OutputMimeType: "text/plain",
	},
	FunctionRenderURLToPDF: {
		Name:           FunctionRenderURLToPDF,
		Image:          "gonitro/unoconv2",
		Command:        []string{"sh", "-c", "render-url-to-pdf"},
		TimeoutMS:      60000,
		OutputMimeType: "application/pdf",
	},
	FunctionConvertMarkdown: {
		Name:           FunctionConvertMarkdown,
		Image:          "busybox",
		Command:        []string{"sh", "-c", "convert-to-markdown"},
		TimeoutMS:      30000,
		OutputMimeType: "text/markdown",
	},
}

func LookupFunctionSpec(functionName string) (FunctionSpec, bool) {
	spec, ok := FunctionRegistry[functionName]
	return spec, ok
}

func ValidateUppercaseInput(in UppercaseInput) error {
	if strings.TrimSpace(in.Text) == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

func ValidateURLToPDFInput(in URLToPDFInput) error {
	if err := validateHTTPURL(in.URL); err != nil {
		return err
	}
	return nil
}

func ValidateMarkdownInput(in MarkdownInput) error {
	hasURL := strings.TrimSpace(in.URL) != ""
	hasHTML := strings.TrimSpace(in.HTML) != ""

	if hasURL == hasHTML {
		return fmt.Errorf("exactly one of url or html is required")
	}

	if hasURL {
		if err := validateHTTPURL(in.URL); err != nil {
			return err
		}
	}

	return nil
}

func validateHTTPURL(raw string) error {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("url must be a valid absolute URL")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}

	if strings.TrimSpace(u.Host) == "" {
		return fmt.Errorf("url host is required")
	}

	return nil
}
