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
~/venv/bin/python3 build_schedule.py "Juan Perez" 06082026

# Filled-in sheets from a captured payload
~/venv/bin/python3 build_schedule.py --json payload.json --outdir out/
```

reportlab lives in `~/venv` on the home server, not in the system python, which
is why `python` in the config is that interpreter's full path.

Every row can carry values: the fourteen day rows, the `TOTAL` row, and the
spare rows under it. A day with more than one punch puts its first punch on
that day's own row and spills the rest into the spare rows, which grow past
the usual four if they have to. A punch that is still open prints its
`ENTRADA` with `SALIDA` and `HORAS` blank, and is left out of `TOTAL` — a
printed sheet shouldn't claim hours for a shift that hasn't ended.

It writes one PDF per worker and prints each path on stdout, which is how the
server learns what was produced.

### Tests

```sh
~/venv/bin/python3 -m unittest discover -s schedule_listener   # from the repo root
```

`test_build_schedule.py` checks the cell layout through `sheet_rows()` and then
builds real PDFs and reads them back with `pdftotext` (those cases skip if
poppler is missing). `go test ./internal/schedule/...` covers the other half —
turning punches into the payload these tests take as their input.

## Deploying

The server is part of the klapp Go module (it imports `internal/schedule`), so
it is built from the repo root either way.

The repo lives on the home server itself, so the usual install is local —
`install.sh` does the whole thing (it asks for sudo twice, for `/opt` and for
the unit):

```sh
schedule_listener/install.sh
```

From a different machine, build and copy instead:

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

reportlab has to be present in the venv the config points at:

```sh
~/venv/bin/pip install reportlab
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

`print_command` is an argv the listener runs once per generated PDF with the
file's path appended. On the home server that is CUPS:

```json
"print_command": ["lp", "-d", "HP_OfficeJet_9120e"]
```

`HP_OfficeJet_9120e` is the queue that talks to the printer directly
(`ipp://HP6C0B5ED86EF6.local:631/ipp/print`); `lpstat -p` lists the rest, which
are the driverless duplicates CUPS discovered for the same machine. An empty
array turns printing off and the job stops at "PDF on disk".

A failure there is logged and reported back to the admin as a warning — the
PDFs still exist and can be printed by hand.

To exercise the whole chain without spending paper, hold the jobs in the queue:

```json
"print_command": ["lp", "-d", "HP_OfficeJet_9120e", "-H", "hold"]
```

```sh
lpstat -o                  # the held jobs
cancel HP_OfficeJet_9120e-81
```
