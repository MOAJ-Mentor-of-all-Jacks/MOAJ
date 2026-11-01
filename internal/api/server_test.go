package api

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/config"
)

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

type pipeListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newPipeListener() *pipeListener {
	return &pipeListener{
		conns:  make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return pipeAddr{} }

func (l *pipeListener) dialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	clientConn, serverConn := net.Pipe()
	select {
	case l.conns <- serverConn:
		return clientConn, nil
	case <-l.closed:
		_ = clientConn.Close()
		_ = serverConn.Close()
		return nil, net.ErrClosed
	case <-ctx.Done():
		_ = clientConn.Close()
		_ = serverConn.Close()
		return nil, ctx.Err()
	}
}

func TestServerTimeoutConfiguration(t *testing.T) {
	cfg := config.Config{HTTPAddr: "127.0.0.1:8000", ReadTimeout: time.Second, ReadHeaderTimeout: 2 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 4 * time.Second}
	srv := NewHTTPServer(cfg, http.NotFoundHandler(), testOptions(io.Discard).Logger)
	if srv.Addr != cfg.HTTPAddr || srv.ReadTimeout != cfg.ReadTimeout || srv.ReadHeaderTimeout != cfg.ReadHeaderTimeout || srv.WriteTimeout != cfg.WriteTimeout || srv.IdleTimeout != cfg.IdleTimeout {
		t.Fatal("server dropped configured address or timeouts")
	}
}

func TestServeDrainsInflightRequest(t *testing.T) {
	listener := newPipeListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, release, draining := make(chan struct{}), make(chan struct{}), make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-release:
			_, _ = w.Write([]byte("finished"))
		case <-r.Context().Done():
			return
		}
	})}
	t.Cleanup(func() { _ = server.Close() })
	served := make(chan error, 1)
	go func() {
		served <- Serve(ctx, server, listener, 3*time.Second, func() { close(draining) }, testOptions(io.Discard).Logger)
	}()
	response := make(chan string, 1)
	go func() {
		client := &http.Client{
			Timeout:   5 * time.Second,
			Transport: &http.Transport{DialContext: listener.dialContext},
		}
		res, err := client.Get("http://pipe/test")
		if err != nil {
			response <- err.Error()
			return
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		response <- string(body)
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case <-draining:
	case <-time.After(5 * time.Second):
		t.Fatal("drain hook did not run")
	}
	select {
	case err := <-served:
		t.Fatalf("server stopped before active request completed: %v", err)
	default:
	}
	close(release)
	if body := <-response; body != "finished" {
		t.Fatalf("in-flight response was interrupted: %s", body)
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestServeClosesRequestsAfterShutdownDeadline(t *testing.T) {
	listener := newPipeListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, stopped := make(chan struct{}), make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(stopped)
	})}
	t.Cleanup(func() { _ = server.Close() })
	served := make(chan error, 1)
	go func() {
		served <- Serve(ctx, server, listener, 30*time.Millisecond, func() {}, testOptions(io.Discard).Logger)
	}()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		client := &http.Client{
			Timeout:   5 * time.Second,
			Transport: &http.Transport{DialContext: listener.dialContext},
		}
		res, err := client.Get("http://pipe/test")
		if err == nil {
			_ = res.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	if err := <-served; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want deadline error", err)
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("timed-out request was not canceled")
	}
	<-clientDone
}
