package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func lookupMap(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

func TestDefaults(t *testing.T) {
	c, err := load(lookupMap(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != "127.0.0.1:8000" || c.ReadHeaderTimeout != 5*time.Second || c.MaxBodyBytes != 1048576 {
		t.Fatalf("unexpected safe defaults: %#v", c)
	}
	if c.DatabaseURL != "" || c.LicenseFile != "/etc/moaj/moaj_enterprise.lic" || !strings.HasSuffix(c.LicenseFile, ".lic") {
		t.Fatalf("unsafe database default or unexpected license path: %q", c.LicenseFile)
	}
}

func TestSecretFilesAndLegacyAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "database-secret")
	secret := "postgres://moaj:secret@localhost:5432/moaj"
	if err := os.WriteFile(path, []byte(secret+"\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := load(lookupMap(map[string]string{
		"MOAJ_DATABASE_URL_FILE": path,
		"IMS_HTTP_ADDR":          "localhost:9000",
		"HTTP_READ_TIMEOUT_SEC":  "12",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.DatabaseURL != secret || c.HTTPAddr != "localhost:9000" || c.ReadTimeout != 12*time.Second {
		t.Fatal("secret loading or legacy alias failed")
	}
}

func TestInvalidExplicitConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
		want   string
	}{
		{"malformed duration", map[string]string{"MOAJ_HTTP_READ_TIMEOUT_SEC": "ten"}, "MOAJ_HTTP_READ_TIMEOUT_SEC"},
		{"overflow", map[string]string{"MOAJ_HTTP_READ_TIMEOUT_SEC": "9223372037"}, "MOAJ_HTTP_READ_TIMEOUT_SEC"},
		{"zero", map[string]string{"MOAJ_HTTP_IDLE_TIMEOUT_SEC": "0"}, "MOAJ_HTTP_IDLE_TIMEOUT_SEC"},
		{"empty", map[string]string{"MOAJ_HTTP_ADDR": ""}, "must not be empty"},
		{"unknown env", map[string]string{"MOAJ_ENV": "prod"}, "MOAJ_ENV"},
		{"unknown log", map[string]string{"MOAJ_LOG_LEVEL": "verbose"}, "MOAJ_LOG_LEVEL"},
		{"bad port", map[string]string{"MOAJ_HTTP_ADDR": "localhost:65536"}, "MOAJ_HTTP_ADDR"},
		{"aliases conflict", map[string]string{"MOAJ_HTTP_ADDR": "localhost:9000", "IMS_HTTP_ADDR": "localhost:8000"}, "mutually exclusive"},
		{"legacy aliases conflict", map[string]string{"IMS_HTTP_ADDR": "localhost:9000", "HTTP_ADDR": "localhost:8000"}, "mutually exclusive"},
		{"file conflict", map[string]string{"MOAJ_DATABASE_URL": "postgres://localhost/moaj", "MOAJ_DATABASE_URL_FILE": "secret"}, "mutually exclusive"},
		{"missing file", map[string]string{"MOAJ_METRICS_TOKEN_FILE": filepath.Join(t.TempDir(), "absent")}, "could not be read"},
		{"body limit", map[string]string{"MOAJ_HTTP_MAX_BODY_BYTES": "-1"}, "MOAJ_HTTP_MAX_BODY_BYTES"},
		{"public metrics", map[string]string{"MOAJ_HTTP_ADDR": ":8000"}, "MOAJ_METRICS_TOKEN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load(lookupMap(tc.values))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestProductionRequiresIdentityAndMetricsProtection(t *testing.T) {
	values := map[string]string{"MOAJ_ENV": "production"}
	if _, err := load(lookupMap(values)); err == nil {
		t.Fatal("production defaults unexpectedly accepted")
	}
	values["MOAJ_APP_DOMAIN"] = "moaj.example.org"
	values["MOAJ_METRICS_TOKEN"] = strings.Repeat("x", 32)
	values["MOAJ_HTTP_ADDR"] = ":8000"
	if _, err := load(lookupMap(values)); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsDoNotAppearInErrorsOrStructuredLogs(t *testing.T) {
	secret := "secret-that-must-not-be-logged"
	_, err := load(lookupMap(map[string]string{"MOAJ_DATABASE_URL": "://" + secret}))
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("unsafe validation error: %v", err)
	}
	c, err := load(lookupMap(map[string]string{"MOAJ_DATABASE_URL": "postgres://moaj:" + secret + "@localhost/moaj", "MOAJ_METRICS_TOKEN": secret}))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	slog.New(slog.NewJSONHandler(&output, nil)).Info("configuration", "config", c)
	if strings.Contains(output.String(), secret) || strings.Contains(output.String(), "postgres://") {
		t.Fatal("configuration leaked secrets into logs")
	}
}

func TestSecretFileBounds(t *testing.T) {
	for _, data := range []string{"\r\n", strings.Repeat("x", 65537)} {
		path := filepath.Join(t.TempDir(), "secret")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := load(lookupMap(map[string]string{"MOAJ_METRICS_TOKEN_FILE": path})); err == nil {
			t.Fatal("invalid secret file accepted")
		}
	}
}
