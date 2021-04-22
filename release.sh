#!/bin/sh
DATE=$(date -u '+%Y-%m-%d__%H:%M:%S_(%Z)') goreleaser release --rm-dist && \
cp -a ./dist/teamtalk_bot_*.* ./dist/checksums.txt* ~/vrnw/public_html/tt_bot/latest/
