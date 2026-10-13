# Changelog

All notable changes to this project will be documented in this file.

## [Pre-alpha] - 2026-10-13

### Added
- Java 27 Adoptium toolchain integration (`.sdkmanrc`, `build.gradle`).
- Diffplug Spotless plugin (v8.10.4) for automated code style enforcement and formatting.
- Foojay toolchain resolver convention in `settings.gradle`.

### Changed
- Upgraded Gradle wrapper to version 9.9.0-milestone-2.
- Upgraded project dependencies:
  - `org.apache.logging.log4j:log4j-core` to `2.26.1`
  - `org.slf4j:slf4j-api` and `org.slf4j:slf4j-reload4j` to `2.0.18`
  - `com.google.guava:guava` to `32.0.1-jre`
  - `com.google.code.gson:gson` to `2.8.9`
- Modernized GameScript builder task (`buildGameScript`) to invoke `python3`.
- Modernized Java source code with pattern matching for `instanceof` and improved type safety.
- Re-formatted entire codebase according to Spotless Java formatting rules.

### Fixed
- Fixed typo in disconnection log messages ("disconneting" -> "disconnecting").
