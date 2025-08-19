# modup

[![Release](https://img.shields.io/github/v/release/chaindead/modup?include_prereleases&sort=semver)](https://github.com/chaindead/modup/releases)
[![Build](https://github.com/chaindead/modup/actions/workflows/release.yml/badge.svg)](https://github.com/chaindead/modup/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/chaindead/modup.svg)](https://pkg.go.dev/github.com/chaindead/modup)
[![Go Report Card](https://goreportcard.com/badge/github.com/chaindead/modup)](https://goreportcard.com/report/github.com/chaindead/modup)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)
![Visitors](https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2Fchaindead%2Fmodup&label=Visitors&labelColor=%23d9e3f0&countColor=%23697689&style=flat&labelStyle=none)

## What is it?

Clean terminal UI that scans your Go modules and helps you update selected dependencies intentionally. Built with Bubble Tea, it's responsive, fast, and pleasant to use right in your terminal.

- Scans dependencies and shows where updates are available.
- Lets you pick exactly which modules to update.
- Applies updates one by one with clear, visual progress.

Just [install](#usage) and run `modup` in project root

![golangci-lint repo demo](./examples/demo.gif)

## Install

### Homebrew

```bash
brew install chaindead/homebrew-tap/modup
```

### Linux

<details>

> **Note:** The commands below install to `/usr/local/bin`. To install elsewhere, replace `/usr/local/bin` with your preferred directory in your PATH.

First, download the archive for your architecture:

```bash
# For x86_64 (64-bit)
curl -L -o modup.tar.gz https://github.com/chaindead/modup/releases/latest/download/modup_Linux_x86_64.tar.gz

# For ARM64
curl -L -o modup.tar.gz https://github.com/chaindead/modup/releases/latest/download/modup_Linux_arm64.tar.gz
```

Then install the binary:

```bash
# Extract the binary
sudo tar xzf modup.tar.gz -C /usr/local/bin

# Make it executable
sudo chmod +x /usr/local/bin/modup

# Clean up
rm modup.tar.gz
```
</details>

### Go Install (from sources)

Requires Golang 1.24+

```bash
go install github.com/chaindead/modup@latest
```

## Usage

Run inside a Go project:

```bash
modup
```

## Alternatives

- https://github.com/oligot/go-mod-upgrade — interactive module updates via browser/CLI
- https://github.com/psampaz/go-mod-outdated — report of outdated modules
- https://github.com/icholy/gomajor — discover available major updates
- Renovate Bot, Dependabot — automated dependency update PRs
- Built-ins: `go list -m -u all`, `go get -u` — basic but less controlled

