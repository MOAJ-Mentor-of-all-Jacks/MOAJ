// Package config loads process settings without exposing secrets in errors.
package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is shared by the server and future worker processes.
// Database connectivity is not required by the Phase 1 HTTP skeleton.
type Config struct {
	Env               string
	LogLevel          slog.Level
	AppDomain         string
	HTTPAddr          string
	DatabaseURL       string
	ValkeyAddr        string
	LicenseFile       string
	MetricsToken      string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxBodyBytes      int64
}

// Load reads MOAJ_* variables and documented legacy aliases (IMS_* and pre-prefix
// names). Direct values, aliases, and _FILE variants are mutually exclusive.
// Empty explicit values and malformed settings are errors, rather than silent defaults.
func Load() (Config, error) { return load(os.LookupEnv) }

func load(lookup func(string) (string, bool)) (Config, error) {
	var problems []error
	read := func(fallback string, names ...string) string {
		value, err := readValue(lookup, true, fallback, names...)
		if err != nil {
			problems = append(problems, err)
		}
		return value
	}
	c := Config{
		Env:          read("development", "MOAJ_ENV", "IMS_ENV"),
		AppDomain:    read("localhost", "MOAJ_APP_DOMAIN", "IMS_APP_DOMAIN", "APP_DOMAIN"),
		HTTPAddr:     read("127.0.0.1:8000", "MOAJ_HTTP_ADDR", "IMS_HTTP_ADDR", "HTTP_ADDR"),
		DatabaseURL:  read("", "MOAJ_DATABASE_URL", "IMS_DATABASE_URL", "DATABASE_URL"),
		ValkeyAddr:   read("localhost:6379", "MOAJ_VALKEY_ADDR", "IMS_VALKEY_ADDR", "REDIS_ADDR"),
		MetricsToken: read("", "MOAJ_METRICS_TOKEN", "IMS_METRICS_TOKEN"),
	}
	var err error
	// This setting is itself a path; do not interpret its value as file contents.
	c.LicenseFile, err = readValue(lookup, false, "/etc/moaj/moaj_enterprise.lic", "MOAJ_LICENSE_FILE", "IMS_LICENSE_FILE", "MINTERN_LICENSE_FILE")
	if err != nil {
		problems = append(problems, err)
	}
	switch strings.ToUpper(read("INFO", "MOAJ_LOG_LEVEL", "IMS_LOG_LEVEL")) {
	case "DEBUG":
		c.LogLevel = slog.LevelDebug
	case "INFO":
		c.LogLevel = slog.LevelInfo
	case "WARN":
		c.LogLevel = slog.LevelWarn
	case "ERROR":
		c.LogLevel = slog.LevelError
	default:
		problems = append(problems, errors.New("MOAJ_LOG_LEVEL must be DEBUG, INFO, WARN, or ERROR"))
	}
	duration := func(fallback string, names ...string) time.Duration {
		seconds, err := strconv.ParseInt(read(fallback, names...), 10, 64)
		const maxSeconds = int64((1<<63 - 1) / time.Second)
		if err != nil || seconds <= 0 || seconds > maxSeconds {
			problems = append(problems, fmt.Errorf("%s must be a positive integer number of seconds within time.Duration range", names[0]))
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	c.ReadTimeout = duration("10", "MOAJ_HTTP_READ_TIMEOUT_SEC", "IMS_HTTP_READ_TIMEOUT_SEC", "HTTP_READ_TIMEOUT_SEC")
	c.ReadHeaderTimeout = duration("5", "MOAJ_HTTP_READ_HEADER_TIMEOUT_SEC", "IMS_HTTP_READ_HEADER_TIMEOUT_SEC")
	c.WriteTimeout = duration("15", "MOAJ_HTTP_WRITE_TIMEOUT_SEC", "IMS_HTTP_WRITE_TIMEOUT_SEC", "HTTP_WRITE_TIMEOUT_SEC")
	c.IdleTimeout = duration("60", "MOAJ_HTTP_IDLE_TIMEOUT_SEC", "IMS_HTTP_IDLE_TIMEOUT_SEC", "HTTP_IDLE_TIMEOUT_SEC")
	c.ShutdownTimeout = duration("10", "MOAJ_HTTP_SHUTDOWN_TIMEOUT_SEC", "IMS_HTTP_SHUTDOWN_TIMEOUT_SEC", "HTTP_SHUTDOWN_TIMEOUT_SEC")
	c.MaxBodyBytes, err = strconv.ParseInt(read("1048576", "MOAJ_HTTP_MAX_BODY_BYTES", "IMS_HTTP_MAX_BODY_BYTES"), 10, 64)
	if err != nil {
		problems = append(problems, errors.New("MOAJ_HTTP_MAX_BODY_BYTES must be a positive integer"))
	}
	if err := c.Validate(); err != nil {
		problems = append(problems, err)
	}
	if err := errors.Join(problems...); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Validate also protects callers that construct Config without Load.
func (c Config) Validate() error {
	var problems []error
	if c.Env != "development" && c.Env != "test" && c.Env != "production" {
		problems = append(problems, errors.New("MOAJ_ENV must be development, test, or production"))
	}
	if c.LogLevel != slog.LevelDebug && c.LogLevel != slog.LevelInfo && c.LogLevel != slog.LevelWarn && c.LogLevel != slog.LevelError {
		problems = append(problems, errors.New("MOAJ_LOG_LEVEL must be DEBUG, INFO, WARN, or ERROR"))
	}
	if c.AppDomain == "" || strings.ContainsAny(c.AppDomain, "/?#@ \t\r\n") {
		problems = append(problems, errors.New("MOAJ_APP_DOMAIN must be a host name without scheme, path, or credentials"))
	}
	host, port, err := net.SplitHostPort(c.HTTPAddr)
	if err != nil || !validPort(port) {
		problems = append(problems, errors.New("MOAJ_HTTP_ADDR must be a host:port address with port 1-65535"))
	}
	if host, port, err := net.SplitHostPort(c.ValkeyAddr); err != nil || host == "" || !validPort(port) {
		problems = append(problems, errors.New("MOAJ_VALKEY_ADDR must be a host:port address with port 1-65535"))
	}
	if c.DatabaseURL != "" {
		u, err := url.Parse(c.DatabaseURL)
		if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
			// url.Parse errors may include password-bearing input; do not wrap them.
			problems = append(problems, errors.New("MOAJ_DATABASE_URL must be a valid postgres:// or postgresql:// URL with a host"))
		}
	}
	if c.LicenseFile == "" {
		problems = append(problems, errors.New("MOAJ_LICENSE_FILE must not be empty"))
	}
	for _, setting := range []struct {
		name  string
		value time.Duration
	}{
		{"MOAJ_HTTP_READ_TIMEOUT_SEC", c.ReadTimeout},
		{"MOAJ_HTTP_READ_HEADER_TIMEOUT_SEC", c.ReadHeaderTimeout},
		{"MOAJ_HTTP_WRITE_TIMEOUT_SEC", c.WriteTimeout},
		{"MOAJ_HTTP_IDLE_TIMEOUT_SEC", c.IdleTimeout},
		{"MOAJ_HTTP_SHUTDOWN_TIMEOUT_SEC", c.ShutdownTimeout},
	} {
		if setting.value <= 0 {
			problems = append(problems, fmt.Errorf("%s must be positive", setting.name))
		}
	}
	if c.MaxBodyBytes <= 0 {
		problems = append(problems, errors.New("MOAJ_HTTP_MAX_BODY_BYTES must be positive"))
	}
	if strings.ContainsAny(c.MetricsToken, " \t\r\n") {
		problems = append(problems, errors.New("MOAJ_METRICS_TOKEN must not contain whitespace"))
	}
	loopback := host == "localhost"
	if ip := net.ParseIP(host); ip != nil {
		loopback = ip.IsLoopback()
	}
	if (c.Env == "production" || !loopback) && len(c.MetricsToken) < 32 {
		problems = append(problems, errors.New("MOAJ_METRICS_TOKEN must contain at least 32 characters in production or when listening beyond loopback"))
	}
	if c.Env == "production" && (c.AppDomain == "localhost" || c.AppDomain == "127.0.0.1" || c.AppDomain == "::1") {
		problems = append(problems, errors.New("MOAJ_APP_DOMAIN must identify the deployment in production"))
	}
	return errors.Join(problems...)
}

// LogValue omits credentials and secret paths from structured logs.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(slog.String("environment", c.Env), slog.String("log_level", c.LogLevel.String()),
		slog.String("http_addr", c.HTTPAddr), slog.Bool("database_configured", c.DatabaseURL != ""),
		slog.Bool("metrics_auth_configured", c.MetricsToken != ""))
}

func validPort(value string) bool {
	port, err := strconv.Atoi(value)
	return err == nil && port > 0 && port <= 65535
}

func readValue(lookup func(string) (string, bool), files bool, fallback string, names ...string) (string, error) {
	var selectedName, selectedValue string
	fromFile := false
	for _, name := range names {
		keys := []string{name}
		if files {
			keys = append(keys, name+"_FILE")
		}
		for _, key := range keys {
			if value, ok := lookup(key); ok {
				if selectedName != "" {
					return "", fmt.Errorf("%s and %s are mutually exclusive", selectedName, key)
				}
				if strings.TrimSpace(value) == "" {
					return "", fmt.Errorf("%s must not be empty", key)
				}
				selectedName, selectedValue, fromFile = key, value, key != name
			}
		}
	}
	if selectedName == "" {
		return fallback, nil
	}
	if !fromFile {
		return selectedValue, nil
	}
	file, err := os.Open(selectedValue)
	if err != nil {
		return "", fmt.Errorf("%s could not be read", selectedName)
	}
	defer file.Close()
	const maxSecretBytes = 64 * 1024
	data, err := io.ReadAll(io.LimitReader(file, maxSecretBytes+1))
	if err != nil || len(data) > maxSecretBytes {
		return "", fmt.Errorf("%s must be readable and at most 64 KiB", selectedName)
	}
	value := strings.TrimRight(string(data), "\r\n")
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s contains an empty value", selectedName)
	}
	return value, nil
}
