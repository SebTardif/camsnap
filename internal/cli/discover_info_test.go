package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/steipete/camsnap/internal/discovery"
)

func TestDiscoverInfoUsesALiveContext(t *testing.T) {
	const body = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <tds:GetDeviceInformationResponse>
      <tds:Manufacturer>Acme</tds:Manufacturer>
      <tds:Model>Cam100</tds:Model>
      <tds:FirmwareVersion>1.2.3</tds:FirmwareVersion>
    </tds:GetDeviceInformationResponse>
  </s:Body>
</s:Envelope>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	old := discoverDevices
	discoverDevices = func(ctx context.Context, _ time.Duration) ([]discovery.Device, error) {
		<-ctx.Done()
		return []discovery.Device{{Address: srv.URL, Host: "192.0.2.10:80"}}, nil
	}
	t.Cleanup(func() { discoverDevices = old })

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRootCommand("test")
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"discover", "--info", "--timeout", "40ms"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Cam100") {
		t.Fatalf("stdout = %q; want model from a live info request", stdout.String())
	}
}
