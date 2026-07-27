// Vendored from github.com/batuergun/ubiquiti-mobile-router-proxy/umr.
// Re-copy intentionally if upstream changes; do not import as a Go module.

// Package umr is a Go client for the UniFi Mobile Router (UMR Ultra) local
// portal's undocumented JSON-RPC API.
package umr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// localUsername is the literal username the UMR firmware accepts. The
// portal hardcodes it; only the owner password varies.
const localUsername = "ui"

// Default session timeout requested at login: ~24.6 days, matching the portal.
const defaultSessionTimeoutSec = 2129920

// Client is a thread-safe UMR API client. Construct via New, authenticate
// via Login. Once logged in, the typed methods (InfoDump, GetDeviceInfo, ...)
// transparently attach the session token and re-login once on AccessDenied.
type Client struct {
	host        string
	httpClient  *http.Client
	autoRelogin bool

	mu       sync.Mutex
	session  *session
	password string // retained for transparent re-login
}

// Option configures a Client.
type Option func(*Client)

// WithInsecureTLS disables TLS certificate verification. UMR units use a
// self-signed cert by default, which is appropriate on a LAN.
func WithInsecureTLS() Option {
	return func(c *Client) {
		c.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
}

// WithTimeout sets the per-request HTTP timeout. Default: 10s.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithHTTPClient injects a custom *http.Client. Overrides WithInsecureTLS
// and WithTimeout if also set.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithoutAutoRelogin disables the single re-login retry on AccessDenied.
// Use in tests where you want auth failures to surface immediately.
func WithoutAutoRelogin() Option {
	return func(c *Client) { c.autoRelogin = false }
}

// New constructs a Client targeting the given UMR base URL (e.g.
// "https://192.168.105.1"). Trailing slashes are stripped.
func New(host string, opts ...Option) *Client {
	for len(host) > 0 && host[len(host)-1] == '/' {
		host = host[:len(host)-1]
	}
	c := &Client{
		host:        host,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		autoRelogin: true,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Login authenticates with the device-owner password. The session token is
// cached on the Client and used by all subsequent calls.
func (c *Client) Login(ctx context.Context, password string) error {
	c.mu.Lock()
	c.password = password
	c.mu.Unlock()
	return c.loginLocked(ctx)
}

func (c *Client) loginLocked(ctx context.Context) error {
	c.mu.Lock()
	pw := c.password
	c.mu.Unlock()
	if pw == "" {
		return errors.New("umr: Login not called (no password set)")
	}
	body := map[string]any{
		"username": localUsername,
		"password": pw,
		"timeout":  defaultSessionTimeoutSec,
	}
	var sess session
	if err := c.rpc(ctx, "/ubus/call/session", "login", body, "", &sess); err != nil {
		return fmt.Errorf("umr: login: %w", err)
	}
	if sess.Token == "" {
		return errors.New("umr: login: empty session token in response")
	}
	c.mu.Lock()
	c.session = &sess
	c.mu.Unlock()
	return nil
}

// Logout invalidates the current session on the device. Idempotent.
func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	tok := ""
	if c.session != nil {
		tok = c.session.Token
	}
	c.session = nil
	c.mu.Unlock()
	if tok == "" {
		return nil
	}
	return c.rpc(ctx, "/ubus/call/session", "destroy", map[string]any{}, tok, nil)
}

// rawCall performs a uimqtt-namespace call and decodes result into out.
// out may be nil if the caller doesn't need the response.
func (c *Client) rawCall(ctx context.Context, method string, params any, out any) error {
	c.mu.Lock()
	tok := ""
	if c.session != nil {
		tok = c.session.Token
	}
	c.mu.Unlock()
	if tok == "" {
		return errors.New("umr: not logged in (call Login first)")
	}
	err := c.rpc(ctx, "/ubus/call/uimqtt", method, params, tok, out)
	if err != nil && c.autoRelogin && IsAccessDenied(err) {
		if relErr := c.loginLocked(ctx); relErr == nil {
			c.mu.Lock()
			tok = c.session.Token
			c.mu.Unlock()
			err = c.rpc(ctx, "/ubus/call/uimqtt", method, params, tok, out)
		}
	}
	return err
}

// rpc performs one HTTP POST against the UMR's JSON-RPC shim. token may be
// "" for the initial login call. out may be nil to discard the result.
func (c *Client) rpc(ctx context.Context, path, method string, params any, token string, out any) error {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return fmt.Errorf("encode body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	// The portal sends Content-Type: application/x-www-form-urlencoded
	// deliberately so the browser skips the CORS preflight. The body is
	// JSON regardless. The device requires this exact header.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *RPCError       `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode response (status %d): %w; body=%s", resp.StatusCode, err, truncate(string(raw), 200))
	}
	if env.Error != nil {
		return env.Error
	}
	if out != nil && len(env.Result) > 0 {
		if err := json.Unmarshal(env.Result, out); err != nil {
			return fmt.Errorf("decode result for %s: %w", method, err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
