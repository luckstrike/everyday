# Try Everyday on a Raspberry Pi

[← Back to the README](../README.md)

You don't need a touchscreen or a finished wall mount. The first goal is simply **a Pi on your desk showing Everyday on a monitor**.

> **Status:** this is an initial manual guide. Everyday has been developed and checked on macOS; these steps haven't yet been verified on a physical Pi. This isn't a tested unattended appliance installer.

## What you'll need

- A Raspberry Pi 4, the project's intended hardware target.
- A suitable power supply and microSD card with Raspberry Pi OS.
- A landscape HDMI monitor or TV and the right display cable/adapter for your Pi.
- A keyboard and mouse for setup. Touch is optional.
- Internet access for installing tools, live weather, and optional Google Calendar.

Start with hardware you own. Check readability at the native resolution before buying a dedicated screen; fewer forecast/event cards on small displays is intentional.

## 1. Prepare the desktop

Use Raspberry Pi Imager to install **Raspberry Pi OS (64-bit) with a desktop**, not Lite, so the Pi can run a local browser. Configure networking, the timezone, and landscape display orientation.

The official [getting-started guide](https://www.raspberrypi.com/documentation/computers/getting-started.html) covers imaging, cabling, and first boot. Work directly at the Pi's desktop for this first attempt.

## 2. Install the build tools

In a terminal on the Pi:

```sh
sudo apt update
sudo apt install git make
```

Install **Node.js 22.18+** with npm (Node 24 LTS is a suitable choice) and **Go 1.22+**. Use the official [Node installation options](https://nodejs.org/en/download) and [Go installation guide](https://go.dev/doc/install); choose Linux ARM64 builds where applicable for the 64-bit OS. Don't assume distribution packages satisfy the minimum versions.

```sh
node --version
npm --version
go version
```

A macOS `build/everyday` binary can't simply be copied to the Pi. Build on the Pi to produce a binary for its OS and CPU.

## 3. Build and start Everyday

```sh
git clone https://github.com/luckstrike/everyday.git
cd everyday
make install
make run
```

Leave this terminal running. In **Chromium on the Pi**, open **http://127.0.0.1:8080**.

The Pi runs both server and browser. Your Mac's `127.0.0.1` means your Mac, not the Pi. Pointing the Pi browser at a server on another machine isn't currently supported.

For a no-credentials demo, stop the server with Ctrl-C and use `EVERYDAY_MOCK=1 make run`. Stop the demo and use plain `make run` when ready for real data.

## 4. Make it yours

Use the gear to choose Appearance, add Bills, and set Location. A mouse works for every setting. Choose a theme in this Pi browser; it doesn't inherit the one saved on your Mac.

For Google Calendar, follow the [sign-in guide](reference.md#connect-google-calendar) from a terminal **on the Pi desktop**, with the authorization browser on that same Pi. Sign-in uses a local callback; opening its URL on your Mac won't complete the Pi's local flow.

### Bring over existing data (optional)

You can start fresh instead. To reuse your setup:

1. Stop both dashboard servers before copying files.
2. Locate files using the README's [data directory table](../README.md#your-data-and-backups). On Pi OS the default is `~/.config/everyday/`, unless `XDG_CONFIG_HOME` or per-file overrides change it.
3. Back up any destination files, then transfer the files you need over a trusted channel such as SSH/SCP: `config.json`, `bills.json`, and optionally `google.json`. Don't put them in the repository.
4. Make sure they belong to the user running Everyday. For the default location, restrict the credentials file with `chmod 600 ~/.config/everyday/google.json`.
5. Start the Pi server. Keep `google.json` private: it contains a refresh token, not just a settings preference.

Some files might not exist yet. Copying data doesn't copy the browser theme, install the app, or make a Mac binary runnable on Linux.

## 5. Try fullscreen, then kiosk mode

Start with Chromium's fullscreen action (usually **F11**). Keep the keyboard/mouse nearby while testing.

For a deliberate kiosk session, try this from a terminal in the Pi's graphical desktop **after the server is running**:

```sh
chromium --kiosk http://127.0.0.1:8080
```

Check the executable name on your OS image; it may differ. Save browser work and close existing Chromium windows first. **Alt-F4** closes the kiosk window. Avoid `--incognito` if you want the theme to persist.

Raspberry Pi's [official kiosk tutorial](https://www.raspberrypi.com/tutorials/how-to-use-a-raspberry-pi-in-kiosk-mode/) covers browser startup. Desktop/session mechanisms vary, so this guide doesn't prescribe an untested autostart file.

**Browser autostart isn't enough:** the Go server also needs to start with the right working directory and user-owned config files, and be ready before the page loads. `make run` doesn't configure automatic startup, restart supervision, or boot recovery.

Don't blindly apply a read-only/overlay filesystem recipe: Everyday needs persistent writes for bills, paid marks, saved location, and browser preferences. Plan writable storage and backups first.

## Before you call it a wall appliance

- [ ] It's readable from the intended viewing distance, at its real resolution.
- [ ] The Pi's timezone and clock are correct.
- [ ] The display stays awake; check screen blanking/power settings on your OS image.
- [ ] The dashboard keeps updating overnight and crosses midnight correctly.
- [ ] Weather/calendar recover after a temporary network outage.
- [ ] Bills and location survive a server restart; the theme survives a browser restart.
- [ ] You know how to exit fullscreen and reconnect a keyboard/mouse.
- [ ] You have a private backup and have tested clean shutdown/restart before mounting it.

Use Ctrl-C to stop the server and the OS shutdown command/menu before removing power. Rebooting **doesn't automatically restart Everyday** with these manual instructions.

## Keep it local

There is no login. Leave Go on `127.0.0.1`; don't change it to `0.0.0.0`, forward router ports, or put it behind a public tunnel/proxy. The localhost/origin editing checks aren't authentication. Remote access is a separate feature, not a setup shortcut.
