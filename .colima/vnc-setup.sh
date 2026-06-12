#!/usr/bin/env bash
# vnc-setup.sh — runs INSIDE the Colima/Lima guest (streamed via
# `colima ssh -- bash -s -- <geometry> <password>`). Idempotently installs a
# minimal X + VNC desktop and (re)starts vncserver on :1 (port 5901), bound to
# 0.0.0.0 so Lima forwards it to the Mac's 127.0.0.1:5901.
#
# Auth: a VNC password (VncAuth), NOT -SecurityTypes None — TigerVNC refuses an
# auth-less server on a non-localhost bind without --I-KNOW-THIS-IS-INSECURE.
# The port is only reachable from the host via Lima's user-mode NAT (not the
# LAN), but a password is cheap defense-in-depth for a propagating base-layer
# artifact.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
GEOMETRY=${1:-1600x1000}   # via `bash -s -- <geometry> <password>` from colima.sh
PASSWORD=${2:-colima}

if ! command -v vncserver >/dev/null 2>&1; then
    echo "  installing tigervnc + fluxbox (one-time)..."
    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        tigervnc-standalone-server tigervnc-common fluxbox xterm dbus-x11 >/dev/null
fi

mkdir -p "$HOME/.vnc"
echo "$PASSWORD" | vncpasswd -f > "$HOME/.vnc/passwd"   # -f: plaintext stdin → obfuscated file
chmod 600 "$HOME/.vnc/passwd"
cat > "$HOME/.vnc/xstartup" <<'EOF'
#!/bin/sh
unset SESSION_MANAGER DBUS_SESSION_BUS_ADDRESS
[ -x /usr/bin/dbus-launch ] && eval "$(dbus-launch --sh-syntax)"
exec fluxbox
EOF
chmod +x "$HOME/.vnc/xstartup"

vncserver -kill :1 >/dev/null 2>&1 || true
# VncAuth (default) reads ~/.vnc/passwd; -localhost no → Lima forwards 5901.
vncserver :1 -geometry "$GEOMETRY" -depth 24 -localhost no >/dev/null 2>&1
echo "  vncserver up on :1 (guest 0.0.0.0:5901), geometry $GEOMETRY"
