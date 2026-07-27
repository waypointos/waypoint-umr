# waypoint-umr

A rover connectivity monitor as a Waypoint no-rebuild module. Built on the
Waypoint module SDK (`github.com/waypointos/waypoint/sdk`): the SDK owns
connect, creds, sd_notify, health, and the stats heartbeat; this repo provides
the router's own logic (portal login, polling, uplink derivation) and publishes
the platform's `uplink` capability.

The module polls a Ubiquiti Mobile Router (UMR Ultra) over the router's local
portal and publishes a snapshot of WAN mode, signal, throughput, router system
state, and connected clients. Read-only: the module never writes back to the
router.

- Manifest: [`module.toml`](module.toml)
- Proto: [`protocol/umr.proto`](protocol/umr.proto)

## Layout

- `cmd/waypoint-module-umr`: entrypoint on `wpmodule.Run`.
- `internal/umr`: the portal client (JSON-RPC over self-signed TLS), copied
  from `github.com/batuergun/ubiquiti-mobile-router-proxy`.
- `internal/poller`: one router poll into a `UmrStats` snapshot, plus the
  `UplinkTelemetry` derivation the platform consumes.
- `internal/config`: host, password, and poll interval, all optional.

## Subjects

Everything lives in the module sandbox `waypoint.<rover>.module.umr.`:

| Leaf | Direction | Message |
|---|---|---|
| `stats` | publish | `UmrStats` at `poll_interval_s` (shares the leaf with the SDK heartbeat) |
| `uplink` | publish | `UplinkTelemetry`; `provides = ["uplink"]` makes the agent mirror it onto the platform's `telemetry.uplink` |
| `health.ready` | subscribe | SDK health probe, replies `ok` |

The `stats` leaf carries two message types, so a subscriber must tell a
`UmrStats` snapshot from the SDK's `ModuleStats` heartbeat before decoding; the
panel does that in `dashboard/src/statsFrame.ts`.

## Build and test

The Go build depends on the Waypoint SDK and protocol bindings via local
`replace` directives pointing at a sibling `../waypoint` checkout. The dashboard
has no such dependency and builds from a plain clone (see below).

With `../waypoint` checked out:

```
go build ./...
go test ./...
```

## Manual dev loop against a dev rover

In the Waypoint repo, boot a dev rover (real agent and core in sim mode):

```
make dev-rover
```

In this repo, run the module against that rover (dev NATS is open, so no creds
are minted):

```
make dev
```

Without a reachable router the login loop retries every 30 seconds and no
snapshot is published, so point `host` at a real UMR to see traffic.

## Dashboard

`dashboard/` holds the module's own React UI. It never imports from the
monorepo: tokens and the `Panel` primitive are copied into
`dashboard/src/ui/`, and the generated protobuf bindings are copied into
`dashboard/src/proto/`.

```
cd dashboard
pnpm install
pnpm build
pnpm test
```

pnpm only, never npm or yarn. The build writes one self-contained ESM bundle
that the host loads directly:

| Bundle | Entry | Host surface |
|---|---|---|
| `dashboard/dist/panel.js` | `src/mount.tsx` | rover tab `m-umr`, from `[ui.static]` |

## Image

Run from the repo root. The dashboard panel must be built first, then `make raw`
assembles the signed-ready `.raw` image:

    cd dashboard && pnpm install && pnpm build && cd ..
    make raw

`make raw` builds the arm64 daemon from `./cmd/waypoint-module-umr`, stages the
binary, manifest, systemd unit, and `dashboard/dist/panel.js`, and produces
`dist/umr-<version>.raw`. It needs `squashfs-tools` (`mksquashfs`) on PATH.

## Release

Tag a version and push; CI (`.github/workflows/release.yml`) builds the panel,
builds the `.raw`, signs it with cosign keyless (Sigstore via GitHub Actions
OIDC, no key to manage), and publishes three assets to the GitHub Release:
`umr-<version>.raw`, `umr-<version>.raw.cosign`, and `manifest.json`.

    git tag v0.4.0 && git push --tags

See [SECURITY.md](SECURITY.md) for the cosign verification command and for how
to report a vulnerability.

## Enable on a rover

UMR is deployed through the proxy's module registry (no image rebuild):

1. In the proxy admin UI, register this repo's URL. The proxy fetches the latest
   release, verifies the cosign signature, and pins the release workflow identity
   on first registration.
2. Enable the module on a rover and supply the per-rover configuration:

       host             = "https://192.168.105.1"
       password         = "<owner password>"
       poll_interval_s  = 5

   The UMR portal serves a self-signed certificate, so the module deliberately
   skips TLS verification for this LAN-local connection; the owner password
   never leaves the rover's local network.

The rover's agent reconciles the desired state within seconds: it fetches the
`.raw`, attaches it as a systemd portable service, writes the config and minted
NATS credentials, and starts the unit. Within about 5 seconds the dashboard's
CONNECTION tab appears for that rover.

## License

[Apache-2.0](LICENSE). See [`NOTICE`](NOTICE) for attribution.
