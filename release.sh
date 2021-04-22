#!/bin/sh
DATE=$(date -u '+%Y-%m-%d__%H:%M:%S_(%Z)') goreleaser release --rm-dist
