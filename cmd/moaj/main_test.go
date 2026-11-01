package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestCommandsDoNotLoadConfigOrEchoUnknownInput(t *testing.T) {
	t.Setenv("MOAJ_HTTP_READ_TIMEOUT_SEC", "malformed")
	for _, command := range []string{"help", "version"} {
		var out bytes.Buffer
		if err := run(context.Background(), []string{command}, &out, io.Discard); err != nil || !strings.Contains(out.String(), "moaj") {
			t.Fatalf("%s: output=%q err=%v", command, out.String(), err)
		}
	}
	err := run(context.Background(), []string{"private-secret-argument"}, io.Discard, io.Discard)
	if err == nil || strings.Contains(err.Error(), "private-secret-argument") || !strings.Contains(err.Error(), "moaj") {
		t.Fatalf("unknown command error leaked input or missed usage: %v", err)
	}
}

func TestServeRejectsInvalidConfigurationBeforeListening(t *testing.T) {
	t.Setenv("MOAJ_HTTP_READ_TIMEOUT_SEC", "malformed-private-value")
	err := run(context.Background(), []string{"serve"}, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "configuration") || strings.Contains(err.Error(), "malformed-private-value") {
		t.Fatalf("expected sanitized configuration error: %v", err)
	}
}
