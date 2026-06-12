#!/usr/bin/env bash
# vnc-setup.sh — runs INSIDE the Colima/Lima guest (streamed via
# `colima ssh -- bash -s -- <geometry>`). Idempotently installs a minimal
# X + VNC desktop and (re)starts vncserver on :1 (port 5901), bound to
# 0.0.0.0 so Lima forwards it to the Mac's 127.0.0.1:5901.
#
# Security: -SecurityTypes None is acceptable because the port is only
# reachable from the host through Lima's user-mode network — not the LAN.
# A password is a future extension.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
GEOMETRY=${1:-1600x1000}   # passed via `bash -s -- <geometry>` from colima.sh

if ! command -v vncserver >/dev/null 2>&1; then
    echo "  installing tigervnc + fluxbox (one-time)..."
    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        tigervnc-standalone-server tigervnc-common fluxbox xterm dbus-x11 >/dev/null
fi

mkdir -p "$HOME/.vnc"
cat > "$HOME/.vnc/xstartup" <<'EOF'
#!/bin/sh
unset SESSION_MANAGER DBUS_SESSION_BUS_ADDRESS
[ -x /usr/bin/dbus-launch ] && eval "$(dbus-launch --sh-syntax)"
exec fluxbox
EOF
chmod +x "$HOME/.vnc/xstartup"

vncserver -kill :1 >/dev/null 2>&1 || true
vncserver :1 -geometry "$GEOMETRY" -depth 24 -localhost no -SecurityTypes None >/dev/null 2>&1
echo "  vncserver up on :1 (guest 0.0.0.0:5901), geometry $GEOMETRY"
