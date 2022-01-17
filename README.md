# logwrap

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Doc](https://pkg.go.dev/badge/github.com/pabigot/logwrap.svg)](https://pkg.go.dev/github.com/pabigot/logwrap)
[![Go Report Card](https://goreportcard.com/badge/github.com/pabigot/logwrap)](https://goreportcard.com/report/github.com/pabigot/logwrap)
[![Build Status](https://github.com/pabigot/logwrap/actions/workflows/core.yml/badge.svg)](https://github.com/pabigot/logwrap/actions/workflows/core.yml)
[![Coverage Status](https://coveralls.io/repos/github/pabigot/logwrap/badge.svg)](https://coveralls.io/github/pabigot/logwrap)

Package logwrap provides a very basic abstraction supporting
syslog-style filterable prioritized string messages.  Logger instances
can be created for specific objects or roles, which can specify an
identifier for themselves.

The use case is helper packages which should emit log messages with the
same tool as the application itself.
