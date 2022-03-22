# Change Log

## [Unreleased]

## [v0.1.0] - 2022-03-22

* Refactor interfaces to distinguish ImmutableLogger which is read-only
  from Logger which retains the ability to set identifiers and priority.

* Add ChanLogger that wraps a logger with a safe-for-concurrent-use F()
  that forwards a packaged log instruction through a channel where it
  can be emitted by the underlying logger in a context that's protected
  from other goroutines.

## [v0.0.7] - 2022-03-01

* Add Enables method on Priority, to allow applications to check whether
  a Logger will process a message at a given priority before collecting
  the information that would be in such a message.

## [v0.0.6] - 2022-02-28

* Add LogOwner interface to indicate support for getting/setting the
  priority of an owned logger.
* Add example of using LogMaker and MakePriLogger.

## [v0.0.5] - 2022-01-26

* Fix testing to be automatable.
* Add MakePriLogger to create helper functions with bound logger and
  priority.

## [v0.0.4] - 2022-01-21

* Support flag.Value API to set priorities from application arguments.
* Add golangci-lint to workflows.

## [v0.0.3] - 2022-01-19

* Add Stringer support to Priority as the numeric values are not meaningful.
* Add ParsePriority to identify a Priority by name.
* Document expected default behavior of new Logger instances.

## [v0.0.2] - 2022-01-17

* Non-functional code cleanup, github workflows for testing and
  coverage, basic README and CHANGELOG.

## v0.0.1 - 2022-01-16

* Initial release with all code, lacking Github actions and non-code
  documentation.

[Unreleased]: https://github.com/pabigot/logwrap/compare/main...next
[v0.0.2]: https://github.com/pabigot/logwrap/compare/v0.0.1...v0.0.2
[v0.0.3]: https://github.com/pabigot/logwrap/compare/v0.0.2...v0.0.3
[v0.0.4]: https://github.com/pabigot/logwrap/compare/v0.0.3...v0.0.4
[v0.0.5]: https://github.com/pabigot/logwrap/compare/v0.0.4...v0.0.5
[v0.0.6]: https://github.com/pabigot/logwrap/compare/v0.0.5...v0.0.6
[v0.0.7]: https://github.com/pabigot/logwrap/compare/v0.0.6...v0.0.7
[v0.1.0]: https://github.com/pabigot/logwrap/compare/v0.0.7...v0.1.0
