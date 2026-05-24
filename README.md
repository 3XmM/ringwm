```markdown
# ringwm

**RingWM** is an experimental window manager for X11 that arranges windows in an ellipse (ring) with the ability to rotate, enlarge the central window, and quickly switch between desktops.

> **Note:** All code was written using **vibecoding** and **Gemini** AI.

## ✨ Features

- 🌀 **Ring layout** — Windows are arranged in an elliptical shape with a depth effect (closer windows are larger, further windows are smaller).
- 🔍 **Zoom** — Enlarge the central window to fill the entire screen.
- 🖥️ **Desktops** — 9 virtual desktops with window isolation.
- ⌨️ **Keyboard control** — All actions are performed using hotkeys.
- ⚡ **Lightweight** — Written in Go using XCB, requires no heavy dependencies.

## ⌨️ Keybinds

| Action               | Combination         |
|----------------------|---------------------|
| Open Terminal        | `Win + Enter`       |
| Close Window         | `Win + Q`           |
| Ring Mode            | `Win + R`           |
| Rotate Ring          | `Win + ←` / `Win + →` |
| Zoom Center Window   | `Win + ↑`           |
| Exit Zoom            | `Win + ↓`           |
| Desktops 1–9         | `Win + 1` … `Win + 9` |
| Exit WM              | `Win + Shift + M`   |

## 📦 Build Dependencies

### Debian / Ubuntu / Mint
```bash
sudo apt update
sudo apt install golang-go xorg-dev libx11-dev libxcb1-dev libxcb-util0-dev libxcb-icccm4-dev libxcb-keysyms1-dev libxcb-ewmh-dev libxcb-randr0-dev
```

Fedora / Red Hat / RHEL / CentOS

```bash
sudo dnf install golang libX11-devel libxcb-devel xcb-util-devel xcb-util-wm-devel xcb-util-keysyms-devel xcb-util-ewmh-devel libxcb-randr0-devel
```

Arch Linux / CachyOS / Manjaro

```bash
sudo pacman -S go libx11 libxcb xcb-util xcb-util-wm xcb-util-keysyms xcb-util-ewmh
```

openSUSE

```bash
sudo zypper install go libX11-devel libxcb-devel xcb-util-devel xcb-util-wm-devel xcb-util-keysyms-devel xcb-util-ewmh-devel
```

Void Linux

```bash
sudo xbps-install -S go libX11-devel libxcb-devel xcb-util-devel xcb-util-wm-devel xcb-util-keysyms-devel xcb-util-ewmh-devel
```

NixOS

Add to configuration.nix:

```nix
environment.systemPackages = with pkgs; [
  go
  libX11
  libxcb
  xcb-util
  xcb-util-wm
  xcb-util-keysyms
  xcb-util-ewmh
];
```

Then run sudo nixos-rebuild switch.

🛠️ Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/3XmM/ringwm
   cd ringwm
   ```
2. Build the binary:
   ```bash
   go build
   ```
3. Install it globally:
   ```bash
   sudo cp ringwm /usr/local/bin/
   ```
4. Add exec ringwm to your ~/.xinitrc and run startx.

⚠️ Warning

This window manager is in early development, so it likely has bugs and other issues.

Test at your own risk!

If you want to help with this project, I'd be happy to — contact me on Telegram: @mxeish

🙏 Thanks

· XGB library for providing the X11 interface.
· The ring layout idea was inspired by cyberpunk aesthetics, Lain from the anime Serial Experiments Lain, and the VxWM infinite canvas.
· The code was written with AI assistance (Gemini).

```
