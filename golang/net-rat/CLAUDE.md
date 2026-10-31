# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

NetRat (`netrat`) is a small single-package Go CLI that prints a one-line summary of the machine's network situation, e.g. `ernst.local / 10.10.0.247 => 195.158.3.42 (Uzbekistan)`. It lives inside the larger `programmer-zendo` repo (module `github.com/squillero/programmer-zendo/golang/netrat`, no external dependencies).

## Commands

```sh
go build .            # produces ./netrat
go run . -vv          # run with Info logging (-vvv for Debug)
go vet . && gofmt -l .
go test .                          # all tests (no network access needed)
go test -run TestNodeInfoUpdate .  # a single test; subtests via -run 'TestNodeInfoUpdate/local_only'
```

Tests use `httptest` servers instead of the real services, and `useTempCache(t)` (in `cache_test.go`) redirects `os.UserCacheDir()` to a temp dir so the user's real cache is never touched. `makeTestNodeInfo` (in `net-info_test.go`) builds a `NodeInfo` from a compact spec. Useful flags while developing: `-t N` timeout in seconds (default 2), `-i` invalidate short-lived cached data, `-z` delete the cache file, `-U` cache-update mode (10 s default timeout, only fills in missing Geo/RDAP data, requires an existing cache). `-v`, `-vv`, `-vvv` are rewritten into `-v=N` in `main.go` before `flag` parsing.

## Architecture

**Rats.** Every information source is a "rat": a type implementing `Rat.Squeal(ni *NodeInfo)` plus a `Query*Rats(...)` function that runs as a goroutine and sends rats on a shared `chan Rat`. Sources: hostname, interfaces (`iface-rats.go`, `hw-rats.go`), private IP via UDP dial to well-known resolvers (`pivate-rats.go` — note the filename typo, `reliable-ips.go`), public IP via plain-text HTTP services (`global-rats.go`), geolocation via JSON HTTP services (`geo-rats.go`), and network name/country via RDAP (`rdap.go`).

**Event loop** (`DescribeNode` in `net-info.go`). Launches all `Query*Rats`, then loops on the channel calling `Squeal` — so `NodeInfo` is mutated only from this goroutine and has no locking; keep mutation inside `Squeal`. After each squeal, `NodeInfo.Update()` recomputes the human-readable fields and returns true once the description is complete; the loop then lingers `timeout/10` to collect late answers and cancels the context. `QueryPrivateRats` sends a `TimeOutRat` sentinel when all UDP dials finish; if no non-loopback private IP was found, the loop gives up immediately (offline case).

**Ephemeras** (`ephemeras.go`). All `NodeInfo` data lives in `Ephemeras[T]`: generic maps of key → value with a timestamp per entry. On startup the JSON cache is loaded and `CleanUp` drops stale entries with two cut-offs: 10 minutes for local/egress data (host, ifaces, private/public IPs; `-i` sets this to "now") and 30 days for `Geo` and `NetworkName`. If the cache alone makes `Update()` return true, no network queries are made at all.

**Cache** (`cache.go`). Stored as JSON at `os.UserCacheDir()/net-rat.json`. A cache whose `schema` field differs from `RAT_VERSION` is discarded, so bump `RAT_VERSION` in `main.go` whenever the serialized shape of `NodeInfo` changes. Only exported fields are persisted; the lowercase fields of `NodeInfo` (`host`, `public`, `geo`, `description`, …) are derived by `Update()`.

**Embedded data** (`data/`, via `//go:embed`): `iso3166.json` drives `canonize()`, which normalizes country codes/aliases from different services to a single name; `CIDRs.json` lists known networks whose name replaces the geo label when the public IP falls in them.

**Adding a source:** define the rat type and its `Squeal`, write a `Query*Rats(ctx, output)` that respects `ctx` (use `FetchHttp`/`FetchRaw` in `http.go` for HTTP), and start it in `DescribeNode` (and in `UpdateCache` if it fills long-lived data).

## Conventions

- Follow `../style.md` (Go style for this repo) and `gofmt`.
- Every `.go` file starts with the ASCII-rat banner header carrying the version and copyright; the README banner carries the version too. These are updated by hand and may lag `RAT_VERSION`.
- Logging: `slog` for diagnostics (levels chosen by `-v`), `log` (prefixed 🐀) for banner and fatal errors; only the final description goes to stdout.
