// Package poller maps a umr.InfoDump into the UmrStats proto. The timing loop
// lives in the module entrypoint; poller is pure mapping.
package poller

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	waypointv1 "github.com/waypointos/waypoint/protocol/gen/go/messages"
	"github.com/waypointos/waypoint/protocol/platform/stamp"

	"github.com/waypointos/waypoint-umr/internal/umr"
	umrv1 "github.com/waypointos/waypoint-umr/protocol/gen/go"
)

// Client is the subset of *umr.Client that poller calls. Keeping it an
// interface lets tests substitute a fake.
type Client interface {
	InfoDump(ctx context.Context) (*umr.InfoDump, error)
}

// Snapshot calls InfoDump and translates it to UmrStats. On error, returns
// an offline-shaped UmrStats with Device.Online=false (not an error) so
// the poll loop can still keep publishing.
func Snapshot(ctx context.Context, c Client) (*umrv1.UmrStats, error) {
	stats := &umrv1.UmrStats{T: timestamppb.New(time.Now())}
	dump, err := c.InfoDump(ctx)
	if err != nil {
		stats.Device = &umrv1.UmrStats_Device{Online: false}
		stats.Wan = &umrv1.UmrStats_Wan{}
		return stats, nil
	}
	stats.Device = mapDevice(dump)
	stats.Wan = mapWan(dump)
	stats.System = mapSystem(dump)
	stats.Throughput = mapThroughput(dump)
	stats.Clients = mapClients(dump)
	return stats, nil
}

func mapDevice(d *umr.InfoDump) *umrv1.UmrStats_Device {
	uptime := int32(d.Low.UptimeSeconds)
	fw := d.Low.Firmware
	return &umrv1.UmrStats_Device{
		Firmware: &fw,
		UptimeS:  &uptime,
		Online:   true,
	}
}

func mapWan(d *umr.InfoDump) *umrv1.UmrStats_Wan {
	active := d.Low.NetworkSource
	wan := &umrv1.UmrStats_Wan{Active: &active}
	if d.Low.GeoIP != "" {
		ip := d.Low.GeoIP
		wan.ExternalIp = &ip
	}
	if d.Low.GeoISP != "" {
		isp := d.Low.GeoISP
		wan.ExternalIsp = &isp
	}
	switch active {
	case umr.NetworkSourceWifiWAN:
		ssid := d.Low.WifiWANSSID
		bars := int32(d.Low.WifiWANSignalLevel)
		status := int32(d.Low.WifiWANStatusCode)
		wan.Wifi = &umrv1.UmrStats_WifiWan{Ssid: &ssid, SignalBars: &bars, StatusCode: &status}
	case umr.NetworkSourceLTE:
		op := d.Low.OperatorName
		state := int32(d.High.LTEState)
		band := d.Low.LTEBand
		mode := d.Low.LTEMode
		rssi := int32(d.High.RSSI)
		rsrp := int32(d.High.RSRP)
		rsrq := int32(d.High.RSRQ)
		ipv4 := d.Low.LTEIPv4Address
		wan.Lte = &umrv1.UmrStats_Lte{
			Operator: &op, State: &state, Band: &band, Mode: &mode,
			RssiDbm: &rssi, RsrpDbm: &rsrp, RsrqDb: &rsrq, Ipv4: &ipv4,
		}
	case umr.NetworkSourceWAN:
		wan.Ethernet = &umrv1.UmrStats_Ethernet{}
	}
	return wan
}

func mapSystem(d *umr.InfoDump) *umrv1.UmrStats_System {
	cpu := int32(d.Medium.CPU)
	mem := int32(d.Medium.Memory)
	return &umrv1.UmrStats_System{CpuPct: &cpu, MemoryPct: &mem}
}

func mapThroughput(d *umr.InfoDump) *umrv1.UmrStats_Throughput {
	down := d.Medium.DownloadSpeed
	up := d.Medium.UploadSpeed
	latAvg := int32(d.Medium.LatencyAvgMs)
	latLoss := int32(d.Medium.LatencyPacketLossCount)
	return &umrv1.UmrStats_Throughput{
		DownBps: &down, UpBps: &up,
		LatencyAvgMs: &latAvg, LatencyLossCount: &latLoss,
	}
}

// UplinkFromStats derives the platform UplinkTelemetry from a UmrStats snapshot.
// Ethernet and offline links carry no signal_bars (N/A is first-class).
func UplinkFromStats(s *umrv1.UmrStats) *waypointv1.UplinkTelemetry {
	u := &waypointv1.UplinkTelemetry{T: timestamppb.New(time.Now()), Stamp: stamp.Now()}
	online := s.GetDevice().GetOnline()
	u.Online = &online
	max := uint32(4)
	switch s.GetWan().GetActive() {
	case "wifiwan":
		mode, bars := "wifi", uint32(s.GetWan().GetWifi().GetSignalBars())
		u.Mode, u.SignalBars, u.SignalBarsMax = &mode, &bars, &max
	case "lte":
		mode, bars := "lte", uint32(s.GetWan().GetLte().GetSignalBars())
		u.Mode, u.SignalBars, u.SignalBarsMax = &mode, &bars, &max
	case "wan":
		mode := "ethernet"
		u.Mode = &mode
	}
	return u
}

func mapClients(d *umr.InfoDump) []*umrv1.UmrStats_Client {
	list := d.Client.List()
	out := make([]*umrv1.UmrStats_Client, 0, len(list))
	for _, c := range list {
		name := c.HostName
		ip := c.IP
		mac := c.MAC
		conn := c.Connection
		speed := int32(c.LinkSpeed)
		uptime := int32(c.UptimeSec)
		out = append(out, &umrv1.UmrStats_Client{
			Name: &name, Ip: &ip, Mac: &mac, Connection: &conn,
			LinkSpeedMbps: &speed, UptimeS: &uptime,
		})
	}
	return out
}
