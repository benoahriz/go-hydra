package main

import "testing"

func TestLookupFunctionSpec(t *testing.T) {
	tests := []struct {
		name     string
		function string
		wantOK   bool
	}{
		{name: "uppercase exists", function: FunctionTextUppercase, wantOK: true},
		{name: "pdf exists", function: FunctionRenderURLToPDF, wantOK: true},
		{name: "markdown exists", function: FunctionConvertMarkdown, wantOK: true},
		{name: "unknown missing", function: "unknown.fn", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := LookupFunctionSpec(tt.function)
			if ok != tt.wantOK {
				t.Fatalf("LookupFunctionSpec(%q) ok = %v, want %v", tt.function, ok, tt.wantOK)
			}
			if ok && spec.Name == "" {
				t.Fatalf("LookupFunctionSpec(%q) returned empty spec name", tt.function)
			}
		})
	}
}

func TestValidateUppercaseInput(t *testing.T) {
	tests := []struct {
		name    string
		in      UppercaseInput
		wantErr bool
	}{
		{name: "valid", in: UppercaseInput{Text: "hello"}, wantErr: false},
		{name: "empty", in: UppercaseInput{Text: ""}, wantErr: true},
		{name: "whitespace", in: UppercaseInput{Text: "   "}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUppercaseInput(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateUppercaseInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURLToPDFInput(t *testing.T) {
	tests := []struct {
		name    string
		in      URLToPDFInput
		wantErr bool
	}{
		{name: "valid https", in: URLToPDFInput{URL: "https://example.com"}, wantErr: false},
		{name: "valid http", in: URLToPDFInput{URL: "http://example.com"}, wantErr: false},
		{name: "missing scheme", in: URLToPDFInput{URL: "example.com"}, wantErr: true},
		{name: "invalid scheme", in: URLToPDFInput{URL: "ftp://example.com"}, wantErr: true},
		{name: "missing host", in: URLToPDFInput{URL: "https:///path"}, wantErr: true},
		{name: "empty", in: URLToPDFInput{URL: ""}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLToPDFInput(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateURLToPDFInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMarkdownInput_URLValidation(t *testing.T) {
	tests := []struct {
		name    string
		in      MarkdownInput
		wantErr bool
	}{
		{name: "invalid scheme", in: MarkdownInput{URL: "ftp://example.com"}, wantErr: true},
		{name: "missing host", in: MarkdownInput{URL: "https:///path"}, wantErr: true},
		{name: "valid url", in: MarkdownInput{URL: "https://example.com"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMarkdownInput(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateMarkdownInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
