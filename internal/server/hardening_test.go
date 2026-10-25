package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestHomeOpsResponsesStringInput(t *testing.T) {
	body := map[string]interface{}{"input": "Compare Microsoft and Cursor", "model": "gpt-5.6-sol"}
	transformResponsesRequestBody(body, "gpt-5.6-sol", "low")
	want := []interface{}{map[string]interface{}{"role": "user", "content": "Compare Microsoft and Cursor"}}
	if !reflect.DeepEqual(body["input"], want) {
		t.Fatalf("input was changed: %#v", body["input"])
	}
	if body["store"] != false {
		t.Fatal("Codex requires store=false")
	}
}

func TestHomeOpsPromptNamesUnchanged(t *testing.T) {
	input := "Zed Cline Roo GitHub Copilot Cursor Microsoft"
	if got := replaceNames(input); got != input {
		t.Fatalf("prompt rewritten: %q", got)
	}
}

func TestHomeOpsConfiguredModelsAreNotSubstituted(t *testing.T) {
	for _, model := range []string{"gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra", "gpt-6-sol", "gpt-6-luna"} {
		if got := normalizeModel(model); got != model {
			t.Errorf("%s normalized to %s", model, got)
		}
	}
}

func TestHomeOpsResponsesToolResultPreserved(t *testing.T) {
	input := []interface{}{map[string]interface{}{
		"type": "function_call_output", "call_id": "call_example", "output": "Microsoft and Cursor",
	}}
	body := map[string]interface{}{"input": input}
	transformResponsesRequestBody(body, "gpt-5.6-sol", "low")
	if !reflect.DeepEqual(body["input"], input) {
		t.Fatalf("tool result changed: %#v", body["input"])
	}
}

// The upstream may omit or mislabel the media type even when the body is SSE.
func TestHomeOpsStreamingResponseHeaders(t *testing.T) {
	const stream = "data: {\"type\":\"response.output_text.delta\",\"delta\":\"OK\"}\n\ndata: [DONE]\n\n"
	for _, convert := range []bool{false, true} {
		for _, contentType := range []string{"", "text/plain; charset=utf-8", "text/event-stream", "text/event-stream; charset=utf-8"} {
			t.Run(strconv.FormatBool(convert)+"/"+contentType, func(t *testing.T) {
				writer := httptest.NewRecorder()
				reader := strings.NewReader(stream)
				body := io.NopCloser(homeOpsReadFunc(func(p []byte) (int, error) {
					if !writer.Flushed {
						t.Error("stream headers must be flushed before reading upstream data")
					}
					return reader.Read(p)
				}))
				headers := http.Header{
					"Content-Length": {strconv.Itoa(len(stream))},
					"Cache-Control":  {"public, max-age=3600"},
					"Connection":     {"close"},
				}
				if contentType != "" {
					headers.Set("Content-Type", contentType)
				}
				server := &Server{logger: zerolog.Nop()}
				server.writeResponse(writer, &http.Response{Header: headers, Body: body}, http.StatusOK, "gpt-5.6-sol", convert, !convert)
				response := writer.Result()
				defer response.Body.Close()
				if response.StatusCode != http.StatusOK {
					t.Fatalf("status = %d", response.StatusCode)
				}
				for key, want := range map[string]string{
					"Content-Type":  "text/event-stream; charset=utf-8",
					"Cache-Control": "no-cache", "Connection": "keep-alive", "Content-Length": "",
				} {
					if got := response.Header.Get(key); got != want {
						t.Errorf("%s = %q, want %q", key, got, want)
					}
				}
				if !convert && writer.Body.String() != stream {
					t.Fatal("native Responses SSE payload changed")
				}
				if convert && (!strings.Contains(writer.Body.String(), `"content":"OK"`) || !strings.Contains(writer.Body.String(), "data: [DONE]")) {
					t.Fatal("Chat Completions SSE conversion failed")
				}
			})
		}
	}
}

type homeOpsReadFunc func([]byte) (int, error)

func (f homeOpsReadFunc) Read(p []byte) (int, error) { return f(p) }

func TestHomeOpsUpstreamErrorResponseUnchanged(t *testing.T) {
	for _, convert := range []bool{false, true} {
		for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusTooManyRequests, http.StatusInternalServerError} {
			for _, contentType := range []string{"application/json", "text/plain; charset=utf-8"} {
				t.Run(strconv.FormatBool(convert)+"/"+strconv.Itoa(status)+"/"+contentType, func(t *testing.T) {
					payload := `{"error":{"message":"upstream rejected request"}}`
					if strings.HasPrefix(contentType, "text/plain") {
						payload = "upstream rejected request"
					}
					headers := http.Header{
						"Content-Type": {contentType}, "Content-Length": {strconv.Itoa(len(payload))},
						"Cache-Control": {"no-store"}, "Retry-After": {"7"},
					}
					writer := httptest.NewRecorder()
					server := &Server{logger: zerolog.Nop()}
					server.writeResponse(writer, &http.Response{Header: headers, Body: io.NopCloser(strings.NewReader(payload))}, status, "gpt-5.6-sol", convert, !convert)
					response := writer.Result()
					defer response.Body.Close()
					if response.StatusCode != status || writer.Body.String() != payload || !reflect.DeepEqual(response.Header, headers) {
						t.Fatal("upstream error status, headers or body changed")
					}
				})
			}
		}
	}
}
