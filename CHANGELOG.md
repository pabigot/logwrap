# Change Log

## [Unreleased]

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
