package poller

import (
	"context"
	"errors"
	"testing"

	"github.com/waypointos/waypoint-umr/internal/umr"
)

type fakeClient struct {
	dump *umr.InfoDump
	err  error
}

func (f *fakeClient) InfoDump(_ context.Context) (*umr.InfoDump, error) {
	return f.dump, f.err
}

func TestSnapshot_WifiWAN(t *testing.T) {
	c := &fakeClient{dump: &umr.InfoDump{
		Low: umr.InfoLow{
			NetworkSource:      "wifiwan",
			GeoIP:              "203.0.113.10",
			GeoISP:             "Orange",
			Firmware:           "1.17.9",
			UptimeSeconds:      2542,
			WifiWANSSID:        "test-ap",
			WifiWANSignalLevel: 3,
			WifiWANStatusCode:  1,
		},
		Medium: umr.InfoMedium{CPU: 46, Memory: 36, LatencyAvgMs: 4, DownloadSpeed: 2_100_000, UploadSpeed: 250_000},
	}}
	stats, err := Snapshot(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Wan.GetActive() != "wifiwan" {
		t.Fatalf("active: %v", stats.Wan.GetActive())
	}
	if stats.Wan.Wifi.GetSignalBars() != 3 {
		t.Fatalf("bars: %v", stats.Wan.Wifi.GetSignalBars())
	}
	if stats.System.GetCpuPct() != 46 {
		t.Fatalf("cpu: %v", stats.System.GetCpuPct())
	}
	if !stats.Device.Online {
		t.Fatal("expected online=true")
	}
}

func TestSnapshot_LTE(t *testing.T) {
	c := &fakeClient{dump: &umr.InfoDump{
		Low:  umr.InfoLow{NetworkSource: "lte", LTEMode: "5g-nsa", LTEBand: "n41", OperatorName: "Orange", LTEIPv4Address: "10.1.2.3"},
		High: umr.InfoHigh{LTEState: 1, RSSI: -73, RSRP: -98, RSRQ: -11, SampleCount: 0},
	}}
	stats, err := Snapshot(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Wan.Lte.GetRssiDbm() != -73 || stats.Wan.Lte.GetRsrpDbm() != -98 {
		t.Fatalf("lte: %+v", stats.Wan.Lte)
	}
	if stats.Wan.Wifi != nil {
		t.Fatal("wifi should be nil on lte")
	}
}

func TestSnapshot_ClientReportsOffline(t *testing.T) {
	c := &fakeClient{err: errors.New("router unreachable")}
	stats, err := Snapshot(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Device.Online {
		t.Fatal("expected offline")
	}
	if stats.Wan.Wifi != nil || stats.Wan.Lte != nil {
		t.Fatalf("wan should be empty when offline: %+v", stats.Wan)
	}
}
