package prompt

import (
	"strings"
	"testing"
)

func TestRenderBasicTokens(t *testing.T) {
	ctx := &Context{
		UserName:   "alice",
		HostName:   "workstation",
		CWD:        "/home/alice/projects",
		HomeDir:    "/home/alice",
		LastStatus: "0",
	}

	tests := []struct {
		format   string
		contains string
	}{
		{"{user}", "alice"},
		{"{host}", "workstation"},
		{"{dir}", "projects"},
		{"{path}", "/home/alice/projects"},
		{"{home_path}", "~/projects"},
		{"{last_status}", "0"},
		{"{os}", ""}, // just check it doesn't crash
	}

	for _, tt := range tests {
		result := Render(tt.format, ctx)
		if tt.contains != "" && !strings.Contains(result, tt.contains) {
			t.Errorf("Render(%q) = %q, expected to contain %q", tt.format, result, tt.contains)
		}
	}
}

func TestRenderHomePath(t *testing.T) {
	ctx := &Context{
		CWD:     "/home/alice",
		HomeDir: "/home/alice",
	}
	result := Render("{home_path}", ctx)
	if result != "~" {
		t.Errorf("expected '~', got %q", result)
	}
}

func TestRenderColors(t *testing.T) {
	ctx := &Context{UserName: "bob"}

	result := Render("{fg:red}{user}{reset}", ctx)
	if !strings.Contains(result, "\033[31m") {
		t.Errorf("expected red ANSI code, got %q", result)
	}
	if !strings.Contains(result, "bob") {
		t.Errorf("expected 'bob' in output, got %q", result)
	}
	if !strings.Contains(result, "\033[0m") {
		t.Errorf("expected reset ANSI code, got %q", result)
	}
}

func TestRenderHexColor(t *testing.T) {
	ctx := &Context{}
	result := Render("{fg:#ff5500}test{reset}", ctx)
	if !strings.Contains(result, "\033[38;2;255;85;0m") {
		t.Errorf("expected true-color ANSI code, got %q", result)
	}
}

func TestRenderBgColor(t *testing.T) {
	ctx := &Context{}
	result := Render("{bg:blue}test{reset}", ctx)
	if !strings.Contains(result, "\033[44m") {
		t.Errorf("expected bg blue ANSI code (44), got %q", result)
	}
}

func TestRenderTextStyles(t *testing.T) {
	ctx := &Context{}

	tests := []struct {
		token string
		code  string
	}{
		{"{bold}", "\033[1m"},
		{"{dim}", "\033[2m"},
		{"{italic}", "\033[3m"},
		{"{underline}", "\033[4m"},
	}

	for _, tt := range tests {
		result := Render(tt.token, ctx)
		if result != tt.code {
			t.Errorf("Render(%q) = %q, expected %q", tt.token, result, tt.code)
		}
	}
}

func TestRenderUnknownToken(t *testing.T) {
	ctx := &Context{}
	result := Render("{unknown_thing}", ctx)
	if result != "{unknown_thing}" {
		t.Errorf("expected passthrough, got %q", result)
	}
}

func TestRenderNewline(t *testing.T) {
	ctx := &Context{UserName: "x"}
	result := Render("{user}{newline}$ ", ctx)
	if result != "x\n$ " {
		t.Errorf("expected 'x\\n$ ', got %q", result)
	}
}

func TestRenderDefault(t *testing.T) {
	ctx := &Context{
		UserName: "alice",
		CWD:      "/tmp/test",
	}
	result := Render("", ctx)
	if !strings.Contains(result, "alice") {
		t.Errorf("default prompt should contain username, got %q", result)
	}
	if !strings.Contains(result, "test") {
		t.Errorf("default prompt should contain dir name, got %q", result)
	}
}

func TestRenderCompositePrompt(t *testing.T) {
	ctx := &Context{
		UserName:   "dev",
		HostName:   "server",
		CWD:        "/var/www",
		HomeDir:    "/home/dev",
		LastStatus: "1",
	}

	format := "{fg:green}{user}{reset}@{fg:blue}{host}{reset}:{fg:yellow}{home_path}{reset} [{last_status}] $ "
	result := Render(format, ctx)

	if !strings.Contains(result, "dev") {
		t.Error("missing username")
	}
	if !strings.Contains(result, "server") {
		t.Error("missing hostname")
	}
	if !strings.Contains(result, "/var/www") {
		t.Error("missing path")
	}
	if !strings.Contains(result, "[1]") {
		t.Error("missing last_status")
	}
}
