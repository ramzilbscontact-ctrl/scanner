#!/usr/bin/env bash
# Generate the demo GIF for README + Show HN + PH launch
# Usage: ./scripts/make-demo-gif.sh
#
# Prerequisites:
#   brew install asciinema gifsicle
#   cargo install --git https://github.com/asciinema/agg
#
# Output: demo.gif in current directory.
set -euo pipefail

SCANNER_BIN="${SCANNER_BIN:-./agenzia-scan}"
if [ ! -x "$SCANNER_BIN" ]; then
  echo "Building agenzia-scan..."
  go build -o agenzia-scan ./cmd/agenzia-scan
fi

echo "Recording asciinema session..."

# Create a fake "demo" scenario using a subshell script
cat > /tmp/agenzia-demo.sh << 'EOF'
#!/bin/bash
clear
sleep 0.5
echo "$ agenzia-scan"
sleep 1
./agenzia-scan
sleep 2
EOF
chmod +x /tmp/agenzia-demo.sh

asciinema rec \
  --cols 90 \
  --rows 28 \
  --title "Agenzia Scanner v0.1.0 — 60-second NIS2 compliance scan" \
  --command "/tmp/agenzia-demo.sh" \
  --overwrite \
  demo.cast

echo "Converting to GIF..."
agg \
  --cols 90 \
  --rows 28 \
  --font-size 16 \
  --speed 1.8 \
  --theme monokai \
  demo.cast demo.gif

echo "Optimizing GIF..."
if command -v gifsicle >/dev/null 2>&1; then
  gifsicle -O3 --lossy=80 demo.gif -o demo.gif
fi

size=$(du -h demo.gif | cut -f1)
echo ""
echo "✓ demo.gif ($size)"
echo "  Upload to the repo (or assets CDN) and link from README."

rm -f /tmp/agenzia-demo.sh demo.cast
