#!/bin/sh
set -e
getent group livekit-sip >/dev/null || groupadd --system livekit-sip
getent passwd livekit-sip >/dev/null || useradd --system --gid livekit-sip \
  --home-dir /var/lib/livekit-sip --create-home --shell /usr/sbin/nologin livekit-sip
chown livekit-sip:livekit-sip /var/lib/livekit-sip /etc/livekit-sip/config.yaml
if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload
  systemctl enable livekit-sip >/dev/null 2>&1 || true
fi
