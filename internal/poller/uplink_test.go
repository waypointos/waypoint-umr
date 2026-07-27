package poller

import (
	"testing"

	umrv1 "github.com/waypointos/waypoint-umr/protocol/gen/go"
)

func wanStats(active string, wifiBars, lteBars int32) *umrv1.UmrStats {
	w := &umrv1.UmrStats_Wan{Active: &active}
	w.Wifi = &umrv1.UmrStats_WifiWan{SignalBars: &wifiBars}
	w.Lte = &umrv1.UmrStats_Lte{SignalBars: &lteBars}
	return &umrv1.UmrStats{Device: &umrv1.UmrStats_Device{Online: true}, Wan: w}
}

func TestUplinkFromStats(t *testing.T) {
	u := UplinkFromStats(wanStats("wifiwan", 3, 1))
	if u.GetMode() != "wifi" {
		t.Fatalf("mode: got %q, want wifi", u.GetMode())
	}
	if u.GetSignalBars() != 3 {
		t.Fatalf("signal_bars: got %d, want 3", u.GetSignalBars())
	}
	if u.GetSignalBarsMax() != 4 {
		t.Fatalf("signal_bars_max: got %d, want 4", u.GetSignalBarsMax())
	}
	if !u.GetOnline() {
		t.Fatal("expected online=true")
	}
	if u.GetStamp() == nil || u.GetStamp().GetT() == nil {
		t.Fatalf("stamp: got %v, want a wall-clock stamp", u.GetStamp())
	}

	u = UplinkFromStats(wanStats("lte", 3, 2))
	if u.GetMode() != "lte" {
		t.Fatalf("mode: got %q, want lte", u.GetMode())
	}
	if u.GetSignalBars() != 2 {
		t.Fatalf("signal_bars: got %d, want 2", u.GetSignalBars())
	}

	u = UplinkFromStats(wanStats("wan", 0, 0))
	if u.GetMode() != "ethernet" {
		t.Fatalf("mode: got %q, want ethernet", u.GetMode())
	}
	if u.SignalBars != nil {
		t.Fatal("ethernet carries no bars")
	}

	off := &umrv1.UmrStats{Device: &umrv1.UmrStats_Device{Online: false}, Wan: &umrv1.UmrStats_Wan{}}
	u = UplinkFromStats(off)
	if u.GetOnline() {
		t.Fatal("expected online=false")
	}
	if u.SignalBars != nil {
		t.Fatal("offline carries no bars")
	}
}
