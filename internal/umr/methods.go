package umr

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
)

var clientKeyRE = regexp.MustCompile(`^client(\d+)$`)

// InfoDump returns the full dashboard payload. Prefer this over the
// standalone Info*Dump methods: it returns a strict superset of fields.
func (c *Client) InfoDump(ctx context.Context) (*InfoDump, error) {
	// Decode into a raw map first so we can pull the client0..clientN
	// numbered keys into a stable slice on InfoClient.
	var raw struct {
		Low    json.RawMessage `json:"low"`
		Medium json.RawMessage `json:"medium"`
		High   json.RawMessage `json:"high"`
		GPS    json.RawMessage `json:"gps"`
		Client json.RawMessage `json:"client"`
	}
	if err := c.rawCall(ctx, "InfoDump", map[string]any{}, &raw); err != nil {
		return nil, err
	}
	out := &InfoDump{GPS: raw.GPS}
	if err := unmarshalIfPresent(raw.Low, &out.Low); err != nil {
		return nil, err
	}
	if err := unmarshalIfPresent(raw.Medium, &out.Medium); err != nil {
		return nil, err
	}
	if err := unmarshalIfPresent(raw.High, &out.High); err != nil {
		return nil, err
	}
	if err := decodeInfoClient(raw.Client, &out.Client); err != nil {
		return nil, err
	}
	return out, nil
}

// InfoLowDump returns the abbreviated low-detail view. InfoDump.Low has more
// fields; use that for anything user-facing.
func (c *Client) InfoLowDump(ctx context.Context) (*InfoLow, error) {
	var v InfoLow
	if err := c.rawCall(ctx, "InfoLowDump", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// InfoMediumDump returns throughput + system + sample fields.
func (c *Client) InfoMediumDump(ctx context.Context) (*InfoMedium, error) {
	var v InfoMedium
	if err := c.rawCall(ctx, "InfoMediumDump", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// InfoHighDump returns the abbreviated high-detail view. InfoDump.High has
// more fields; use that for anything user-facing.
func (c *Client) InfoHighDump(ctx context.Context) (*InfoHigh, error) {
	var v InfoHigh
	if err := c.rawCall(ctx, "InfoHighDump", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// InfoClientDump returns the connected-clients view (same shape as
// InfoDump.Client).
func (c *Client) InfoClientDump(ctx context.Context) (*InfoClient, error) {
	var raw json.RawMessage
	if err := c.rawCall(ctx, "InfoClientDump", map[string]any{}, &raw); err != nil {
		return nil, err
	}
	out := &InfoClient{}
	if err := decodeInfoClient(raw, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDeviceInfo returns hardware identity, firmware, MAC, and WAN/LAN IPs.
func (c *Client) GetDeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	var v DeviceInfo
	if err := c.rawCall(ctx, "GetDeviceInfo", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetDeviceStatus reports activation, online state, and (when LTE is the
// active WAN) cellular details.
func (c *Client) GetDeviceStatus(ctx context.Context) (*DeviceStatus, error) {
	var v DeviceStatus
	if err := c.rawCall(ctx, "GetDeviceStatus", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetStatus returns the device's cloud (MQTT) connectivity, not local
// health. Useful for confirming the UMR is reachable from Ubiquiti's cloud.
func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	var v Status
	if err := c.rawCall(ctx, "GetStatus", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetVpnStatus returns the active VPN tunnel state (WireGuard / OpenVPN / IPSec).
func (c *Client) GetVpnStatus(ctx context.Context) (*VpnStatus, error) {
	var v VpnStatus
	if err := c.rawCall(ctx, "GetVpnStatus", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetSimProfile returns the active SIM's profile data. Mostly empty when no
// SIM is inserted.
func (c *Client) GetSimProfile(ctx context.Context) (*SimProfile, error) {
	var v SimProfile
	if err := c.rawCall(ctx, "GetSimProfile", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetRadioPreference returns the configured cellular radio preference
// (band locks, network type lock). Returns nearly-empty when no SIM.
func (c *Client) GetRadioPreference(ctx context.Context) (json.RawMessage, error) {
	var v json.RawMessage
	if err := c.rawCall(ctx, "GetRadioPreference", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// GetRadioCapability returns the modem's supported bands and modes.
// Returns nearly-empty when no SIM/modem populated.
func (c *Client) GetRadioCapability(ctx context.Context) (json.RawMessage, error) {
	var v json.RawMessage
	if err := c.rawCall(ctx, "GetRadioCapability", map[string]any{}, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func unmarshalIfPresent(raw json.RawMessage, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

// decodeInfoClient parses a payload that mixes top-level fields with
// numbered client keys (client0, client1, ...) into our InfoClient shape.
func decodeInfoClient(raw json.RawMessage, out *InfoClient) error {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	// Pull known fields.
	if v, ok := m["clients"]; ok {
		_ = json.Unmarshal(v, &out.Clients)
	}
	if v, ok := m["sample_count"]; ok {
		_ = json.Unmarshal(v, &out.SampleCount)
	}
	if v, ok := m["sample_interval_second"]; ok {
		_ = json.Unmarshal(v, &out.SampleIntervalSec)
	}
	// Collect clientN entries.
	type indexed struct {
		idx   int
		entry ClientEntry
	}
	var entries []indexed
	for k, v := range m {
		match := clientKeyRE.FindStringSubmatch(k)
		if match == nil {
			continue
		}
		i, _ := strconv.Atoi(match[1])
		var ce ClientEntry
		if err := json.Unmarshal(v, &ce); err != nil {
			return err
		}
		entries = append(entries, indexed{i, ce})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].idx < entries[j].idx })
	out.Extra = make(ClientEntries, len(entries))
	for i, e := range entries {
		out.Extra[i] = e.entry
	}
	return nil
}
