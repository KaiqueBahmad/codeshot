package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"codeshot/internal/update"
)

func TestElsewhere(t *testing.T) {
	tests := []struct {
		channel, want string
	}{
		{channelTar, update.Releases},
		{channelZip, update.Releases},
		{"", update.Repo + "#installing"},
	}
	for _, tt := range tests {
		if got := elsewhere(tt.channel); !strings.Contains(got, tt.want) {
			t.Errorf("elsewhere(%q) = %q, want it to point at %s", tt.channel, got, tt.want)
		}
	}
}

func TestUpdateDebLatestAlready(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"tag_name": "v0.5.0", "assets": []}`)
	}))
	t.Cleanup(srv.Close)

	var out bytes.Buffer
	err := updateDeb(update.Source{API: srv.URL, Client: srv.Client()}, "v0.5.0", nil, &out, &out)
	if err != nil {
		t.Fatalf("updateDeb(): %v", err)
	}
	if want := "Using v0.5.0, the latest.\n"; out.String() != want {
		t.Errorf("updateDeb() said %q, want %q", out.String(), want)
	}
}

func TestVersionLine(t *testing.T) {
	defer func(v, c string) { version, channel = v, c }(version, channel)
	version = "v0.5.0"

	channel = ""
	if got := versionLine(); got != "codeshot v0.5.0" {
		t.Errorf("versionLine() = %q, want codeshot v0.5.0", got)
	}
	channel = channelDeb
	if got := versionLine(); got != "codeshot v0.5.0 (deb)" {
		t.Errorf("versionLine() = %q, want codeshot v0.5.0 (deb)", got)
	}
}

func TestUpdateRejectsArguments(t *testing.T) {
	for _, arg := range []string{"now", "--from", "--force"} {
		if got := Main([]string{"update", arg}); got != 2 {
			t.Errorf("update %s exited %d, want 2", arg, got)
		}
	}
}

func TestUpdateOtherChannels(t *testing.T) {
	defer func(c string) { channel = c }(channel)
	for _, c := range []string{"", channelTar, channelZip, "unknown"} {
		channel = c
		var out, stderr bytes.Buffer
		if got := runUpdate(nil, &out, &stderr); got != 1 {
			t.Errorf("channel %q exited %d, want 1", c, got)
		}
		if out.Len() != 0 || stderr.String() != elsewhere(c) {
			t.Errorf("channel %q: stdout %q, stderr %q", c, out.String(), stderr.String())
		}
	}
}
