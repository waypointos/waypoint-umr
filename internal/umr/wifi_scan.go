package umr

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// NearbyAP is one entry in a wifi-scan result. signal_strength uses the same
// coarse 0–4 bar scale as the dashboard; the UMR firmware does not expose
// dBm-level signal data through the local "ui" API.
type NearbyAP struct {
	SSID           string `json:"ssid"`
	SignalStrength int    `json:"signal_strength"` // 0-4 bars
	SecurityMask   int    `json:"security_mask"`   // bitmask; observed: 2 = WPA-PSK
}

// wifi-scan action states. Observed values in the wild: 2 = running,
// 3 = complete, 4 = failed. Other values may exist on future firmwares.
const (
	wifiScanStateRunning  = 2
	wifiScanStateComplete = 3
	wifiScanStateFailed   = 4
)

// WifiScan triggers a wifi scan on the router and waits for results. This
// briefly retunes the radio, so do NOT call it on a polling loop; it is
// intended for on-demand use (e.g. picking a wifi-WAN target).
//
// pollInterval controls how often we poll for completion (typical: 1-2s).
// Returns the list of nearby APs.
func (c *Client) WifiScan(ctx context.Context, pollInterval time.Duration) ([]NearbyAP, error) {
	if pollInterval <= 0 {
		pollInterval = 1500 * time.Millisecond
	}
	var trigger struct {
		OK bool `json:"ok"`
	}
	if err := c.rawCall(ctx, "Action", map[string]any{"action_name": "wifi_scan"}, &trigger); err != nil {
		return nil, fmt.Errorf("trigger wifi_scan: %w", err)
	}
	if !trigger.OK {
		return nil, errors.New("umr: wifi_scan trigger returned ok=false")
	}

	type checkResp struct {
		State      int        `json:"state"`
		NearbyWiFi []NearbyAP `json:"nearby_wifi"`
		OK         bool       `json:"ok"`
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
		var r checkResp
		if err := c.rawCall(ctx, "ActionCheck", map[string]any{"action_name": "wifi_scan"}, &r); err != nil {
			return nil, fmt.Errorf("poll wifi_scan: %w", err)
		}
		switch r.State {
		case wifiScanStateRunning:
			continue
		case wifiScanStateComplete:
			return r.NearbyWiFi, nil
		case wifiScanStateFailed:
			return nil, errors.New("umr: wifi_scan failed (state=4)")
		default:
			return nil, fmt.Errorf("umr: wifi_scan unknown state %d", r.State)
		}
	}
}
