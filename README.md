# logwrap

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Doc](https://pkg.go.dev/badge/github.com/pabigot/logwrap.svg)](https://pkg.go.dev/github.com/pabigot/logwrap)
[![Go Report Card](https://goreportcard.com/badge/github.com/pabigot/logwrap)](https://goreportcard.com/report/github.com/pabigot/logwrap)
[![Build Status](https://github.com/pabigot/logwrap/actions/workflows/core.yml/badge.svg)](https://github.com/pabigot/logwrap/actions/workflows/core.yml)
[![Coverage Status](https://coveralls.io/repos/github/pabigot/logwrap/badge.svg)](https://coveralls.io/github/pabigot/logwrap)

Package logwrap provides a very basic logging abstraction supporting
syslog-style filterable prioritized text messages.  The underlying log
implementation is injected by providing a wrapper object that implements
Logger.  Logger instances can be created for specific objects or roles,
and can specify an identifier for themselves.

Where the underlying log infrastructure is not safe for concurrent use,
MakeChanLogger allows multiple goroutines to send messages through a
channel to a goroutine that exclusively uses the logger.

The use case is helper packages that should emit log messages with the
same tool as the application itself.
