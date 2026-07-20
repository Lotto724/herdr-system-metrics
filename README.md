# herdr-system-metrics

herdr-system-metrics is a local Linux/WSL best-effort popup for current aggregate
CPU, RAM, and swap usage. It is a plugin MVP, not a release.

## Quick path

```sh
go build -o bin/herdr-system-metrics ./cmd/herdr-system-metrics
herdr plugin link "$(pwd)"
herdr plugin pane open --plugin herdr.system-metrics --entrypoint system-metrics
```

The popup is a singleton modal. It closes when the process exits and Herdr returns
`ui_busy` when another modal is active.

## Controls and states

| Input | Result |
| --- | --- |
| `q`, `Esc`, `Ctrl+C` | Close the popup. |
| `space` | Pause or resume collection. |

CPU is `warming up` until two valid samples establish an interval. Missing,
malformed, or invalid metric data is `unavailable`; it is never reported as zero.
A narrow terminal displays `terminal too small` rather than clipping current values
or state-specific controls.

## Linux/WSL semantics

The popup reads only Linux `/proc/stat` and `/proc/meminfo` once per second while
unpaused. RAM uses `MemAvailable` to calculate current used memory. `/proc/loadavg`
is not read or required.
On WSL and in containers, values MAY be VM- or namespace-scoped. They do not
claim to represent the Windows host or another host.

## Non-goals

This MVP does not add temperature, frequency, background collection, persistent
history, storage, alerts, process attribution, remote telemetry, host APIs,
sidebar contributions, or native Metrics. The roadmap validates this popup before
any upstream sidebar proposal or native Metrics consideration.

## Local cleanup

Unlink the local plugin when finished:

```sh
herdr plugin unlink herdr.system-metrics
```
