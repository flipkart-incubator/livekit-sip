#!/bin/sh
set -e
# Debian prerm: $1 is remove|upgrade|deconfigure|failed-upgrade
if [ "$1" = "remove" ] || [ -z "$1" ]; then
  if command -v systemctl >/dev/null 2>&1; then
    systemctl stop livekit-sip >/dev/null 2>&1 || true
    systemctl disable livekit-sip >/dev/null 2>&1 || true
  fi
fi
