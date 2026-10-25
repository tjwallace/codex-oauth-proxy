package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestHomeOpsPrivacyWriter(t *testing.T) {
	var output bytes.Buffer
	writer := privacyWriter{&output}
	input := []byte(`{"level":"warn","message":"Upstream error","status_code":400,"authorization_preview":"secret-token","inbound_body_preview":"secret-prompt","response_body":"secret-output","error":"secret-error","new_upstream_field":"secret-new"}`)
	n, err := writer.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("write: n=%d err=%v", n, err)
	}
	if strings.Contains(output.String(), "secret-") {
		t.Fatalf("sensitive field logged: %s", output.String())
	}
	if !strings.Contains(output.String(), `"status_code":400`) {
		t.Fatal("lost operational status")
	}
	output.Reset()
	_, _ = writer.Write([]byte("invalid record secret-prompt"))
	if output.Len() != 0 {
		t.Fatal("malformed record did not fail closed")
	}
}
