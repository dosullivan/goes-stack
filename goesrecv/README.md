# Pi-side: satdump as a systemd service

This directory holds the artifacts that run on the Raspberry Pi in the yard. The job of the Pi is to demodulate the GOES HRIT signal with [satdump](https://www.satdump.org/) and publish the resulting stream over TCP, so the indoor server (running [goesproc-docker](../goesproc-docker/)) can subscribe to it.

`satdump-goes19.service` is the systemd unit that keeps satdump running indefinitely — it restarts on failure and starts automatically on boot.

## What's in the unit

The included unit targets **GOES-19** on **1694.1 MHz**. All GOES-R series satellites (GOES-16/17/18/19) share the same HRIT downlink frequency, so if you're pointing at GOES-18 instead, just change the satdump pipeline name from `GOES_19` to `GOES_18` (and update `Description` / filename to match) — the frequency stays the same.

It assumes:
- `satdump` is installed at `/usr/bin/satdump` (adjust `ExecStart` if yours is elsewhere — `which satdump` will tell you)
- Your SDR is plugged in and recognized
- You're OK running as `root`. To run unprivileged instead, give the rtlsdr binary the right capabilities (`sudo setcap cap_sys_nice,cap_net_admin+ep $(which satdump)`) and change the `User=` line.

The unit exposes both the satdump web UI on port 8080 and the demodulated TCP stream on port 5004 — `goesproc-docker` subscribes to the latter.

## Installing the unit

On the Pi:

```bash
sudo cp satdump-goes18.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now satdump-goes18.service
```

## Checking that it's running

```bash
systemctl status satdump-goes18           # current state
sudo journalctl -u satdump-goes18 -f      # tail logs
```

You should see satdump locking onto the signal and reporting SNR. Then load `http://<pi-ip>:8080` in a browser to see the live constellation/decoder UI.

## Updating the unit

After editing the file:

```bash
sudo cp satdump-goes18.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl restart satdump-goes18
```
