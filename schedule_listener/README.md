# Schedule listener

Runs on the home server (`10.9.0.7`, reachable over WireGuard). It receives a
pay period's hours from klapp and turns them into printable PDFs.

```
admin site  ──Print button──▶  printsched  ──HTTP POST /print──▶  schedule-listener
                                                                        │
                                                          build_schedule.py (reportlab)
                                                                        │
                                                                  out/*.pdf
```

The JSON on that wire is `schedule.Payload` in `internal/schedule/payload.go`.
Both ends import that one definition, so the shape can't drift; `build_schedule.py`
reads the same fields by name.

## Files

| File | What it is |
|---|---|
| `main.go` | the HTTP server, `POST /print` and `GET /healthz` |
| `build_schedule.py` | draws the PDF; blank sheets or filled-in ones |
| `config.example.json` | copy to `config.json` and edit |
| `schedule-listener.service` | systemd unit |

## Building `build_schedule.py` by hand

```sh
# One blank sheet to fill in with a pen
./build_schedule.py "Juan Perez" 06082026

# Filled-in sheets from a captured payload
./build_schedule.py --json payload.json --outdir out/
```

Every row can carry values: the fourteen day rows, the `TOTAL` row, and the
spare rows under it. A day with more than one punch puts its first punch on
that day's own row and spills the rest into the spare rows, which grow past
the usual four if they have to. A punch that is still open prints its
`ENTRADA` with `SALIDA` and `HORAS` blank, and is left out of `TOTAL` — a
printed sheet shouldn't claim hours for a shift that hasn't ended.

It writes one PDF per worker and prints each path on stdout, which is how the
server learns what was produced.

## Deploying

The server is part of the klapp Go module (it imports `internal/schedule`), so
build it from the repo root and copy the result over:

```sh
GOOS=linux GOARCH=amd64 go build -o /tmp/schedule-listener ./schedule_listener

ssh 10.9.0.7 'sudo mkdir -p /opt/schedule-listener && sudo chown tklaus /opt/schedule-listener'
scp /tmp/schedule-listener schedule_listener/build_schedule.py 10.9.0.7:/opt/schedule-listener/
scp schedule_listener/config.example.json 10.9.0.7:/opt/schedule-listener/config.json
scp schedule_listener/schedule-listener.service 10.9.0.7:/tmp/
ssh 10.9.0.7 'sudo mv /tmp/schedule-listener.service /etc/systemd/system/ &&
              sudo systemctl daemon-reload &&
              sudo systemctl enable --now schedule-listener'
```

reportlab has to be present on the home server:

```sh
ssh 10.9.0.7 'pip3 install --user reportlab'
```

`addr` in the example config is `10.9.0.7:5555` — the WireGuard address
specifically, so the listener is not reachable from the home LAN either. If
the unit starts before WireGuard is up the bind fails and systemd's
`Restart=always` retries until the interface exists.

## Checking it

```sh
curl http://10.9.0.7:5555/healthz          # -> ok
journalctl -u schedule-listener -f
```

From the klapp box, a print without touching the admin UI:

```sh
cd /opt/klapp && ./printsched -period 2026-06-08
```

## Printing

Not wired up yet. `print_command` in the config is the seam: set it to an argv
array and the listener runs it once per generated PDF with the file's path
appended.

```json
"print_command": ["lp", "-d", "office-laser"]
```

A failure there is logged and reported back to the admin as a warning — the
PDFs still exist and can be printed by hand.
