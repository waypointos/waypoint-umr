// Binary waypoint-module-umr polls the UMR Ultra router and publishes UmrStats
// and the uplink rail on the rover's embedded NATS bus.
//
// The SDK owns the module environment contract (rover id, NATS URL, creds); the
// router host, password, and poll interval come from config.toml.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/waypointos/waypoint/sdk/wpmodule"

	"github.com/waypointos/waypoint-umr/internal/config"
	"github.com/waypointos/waypoint-umr/internal/poller"
	"github.com/waypointos/waypoint-umr/internal/umr"
)

const (
	loginRetry     = 30 * time.Second
	requestTimeout = 8 * time.Second
)

func main() {
	// The systemd unit passes --config/--creds/--rover; local runs pass none and
	// rely on the environment the SDK already reads.
	configFlag := flag.String("config", "", "path to config.toml (overrides WAYPOINT_MODULE_CONFIG)")
	credsFlag := flag.String("creds", "", "path to creds.env (used when WAYPOINT_MODULE_CREDS is unset)")
	_ = flag.String("rover", "", "rover id (accepted and ignored; the SDK reads WAYPOINT_ROVER_ID)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	configPath := resolveConfigPath(*configFlag, os.Getenv("WAYPOINT_MODULE_CONFIG"))
	if creds := resolveCredsEnv(*credsFlag, os.Getenv("WAYPOINT_MODULE_CREDS")); creds != "" {
		if err := os.Setenv("WAYPOINT_MODULE_CREDS", creds); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}

	err := wpmodule.Run(context.Background(), wpmodule.Options{ID: "umr"}, func(m *wpmodule.M) error {
		return setup(m, configPath)
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// resolveConfigPath picks the config.toml path: the flag when set, else the
// environment the agent's drop-in may export, else empty for built-in defaults.
func resolveConfigPath(flagVal, env string) string {
	if flagVal != "" {
		return flagVal
	}
	return env
}

// resolveCredsEnv returns the value WAYPOINT_MODULE_CREDS should be set to, or
// empty to leave it alone. The SDK reads only the environment, and an already
// exported value wins so the agent's drop-in stays authoritative.
func resolveCredsEnv(flagVal, env string) string {
	if env != "" {
		return ""
	}
	return flagVal
}

// setup wires the umr module on the SDK runtime. The SDK owns connect, creds,
// sd_notify, health, and the ModuleStats heartbeat; this adds the router poll
// loop feeding the stats and uplink leaves.
func setup(m *wpmodule.M, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	// Login runs off the setup path: the SDK holds sd_notify READY until setup
	// returns, and a cold-booting router can take minutes to accept one.
	go pollLoop(m, cfg)
	return nil
}

// pollLoop retries the router login until it succeeds, then publishes a stats
// and uplink snapshot every cfg.PollInterval until the module shuts down.
func pollLoop(m *wpmodule.M, cfg *config.Config) {
	ctx := context.Background()
	cl := umr.New(cfg.Host, umr.WithInsecureTLS(), umr.WithTimeout(requestTimeout))
	for {
		if err := cl.Login(ctx, cfg.Password); err == nil {
			break
		} else {
			slog.Warn("umr login (will retry)", "err", err)
		}
		select {
		case <-m.Done():
			return
		case <-time.After(loginRetry):
		}
	}

	statsSubject := m.Subject("stats")
	t := time.NewTicker(cfg.PollInterval)
	defer t.Stop()
	for {
		stats, err := poller.Snapshot(ctx, cl)
		if err != nil {
			slog.Warn("umr snapshot", "err", err)
		} else {
			if b, err := proto.Marshal(stats); err == nil {
				_ = m.Publish(statsSubject, b)
			}
			if err := m.PublishUplink(poller.UplinkFromStats(stats)); err != nil {
				slog.Warn("umr uplink publish", "err", err)
			}
		}
		select {
		case <-m.Done():
			return
		case <-t.C:
		}
	}
}
