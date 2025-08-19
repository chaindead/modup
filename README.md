## modup

[![Release](https://img.shields.io/github/v/release/chaindead/modup?include_prereleases&sort=semver)](https://github.com/chaindead/modup/releases)
[![Build](https://github.com/chaindead/modup/actions/workflows/release.yml/badge.svg)](https://github.com/chaindead/modup/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/chaindead/modup.svg)](https://pkg.go.dev/github.com/chaindead/modup)
[![Go Report Card](https://goreportcard.com/badge/github.com/chaindead/modup)](https://goreportcard.com/report/github.com/chaindead/modup)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

### What is it?

Beautiful TUI to scan your project's Go modules for available updates and upgrade selected ones.

![golangci-lint repo demo](./examples/demo.gif)

### Install

#### Homebrew

```bash
brew install chaindead/homebrew-tap/modup

# Update
brew upgrade chaindead/homebrew-tap/modup
```

#### Go Install

```bash
go install github.com/chaindead/modup@latest
```

### Usage

Run inside a Go project:

```bash
modup
```

