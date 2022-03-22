#!/usr/bin/env bash
set -euxo pipefail

# Cleanup to be "stateless" on startup, otherwise pulseaudio daemon can't start
rm -rf /var/run/pulse /var/lib/pulse /root/.config/pulse

# Start pulseaudio as system wide daemon; for debugging it helps to start in non-daemon mode
pulseaudio -D --verbose --exit-idle-time=-1 --system --disallow-exit

# Run RTSP server
RTSP_LOGDESTINATIONS=file ./rtsp-simple-server &

# Run service
XDG_RUNTIME_DIR=$PATH:~/.cache/xdgr go test -v --tags=integration ./test
