#!/bin/sh
set -e
getent group mohini-livekit-sip >/dev/null || groupadd --system mohini-livekit-sip
getent passwd mohini-livekit-sip >/dev/null || useradd --system --gid mohini-livekit-sip \
  --home-dir /var/lib/mohini-livekit-sip --create-home --shell /usr/sbin/nologin mohini-livekit-sip
chown mohini-livekit-sip:mohini-livekit-sip /var/lib/mohini-livekit-sip /etc/mohini-livekit-sip/config.yaml
if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload
  systemctl enable mohini-livekit-sip >/dev/null 2>&1 || true
fi
