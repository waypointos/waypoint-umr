package umr

import (
	"context"
	"encoding/json"
)

// ShadowReadLocal reads the raw shadow envelope for one of the
// ShadowName* constants. For known shadows prefer the typed wrappers.
func (c *Client) ShadowReadLocal(ctx context.Context, name string) (*ShadowResponse, error) {
	var v ShadowResponse
	if err := c.rawCall(ctx, "ShadowReadLocal", map[string]any{"shadow": name}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ShadowReadNetworks returns the typed networks shadow (LAN/WAN/WiFi config,
// configured WAN priority, data usage, etc.).
func (c *Client) ShadowReadNetworks(ctx context.Context) (*ShadowNetworks, error) {
	resp, err := c.ShadowReadLocal(ctx, ShadowNameNetworks)
	if err != nil {
		return nil, err
	}
	var v ShadowNetworks
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, &v); err != nil {
			return nil, err
		}
	}
	return &v, nil
}

// ShadowReadDeviceInfo returns the device-name + GPS toggle shadow.
func (c *Client) ShadowReadDeviceInfo(ctx context.Context) (*ShadowDeviceInfoData, error) {
	resp, err := c.ShadowReadLocal(ctx, ShadowNameDeviceInfo)
	if err != nil {
		return nil, err
	}
	var v ShadowDeviceInfoData
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, &v); err != nil {
			return nil, err
		}
	}
	return &v, nil
}
