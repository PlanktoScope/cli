# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## 0.2.0 - 2023-06-28

- Added logging level command-line flag
- The `dev proc start` subcommand logs segmenter state updates with level info, rather than printing them directly to stdout

## 0.1.0 - 2023-06-28

- Added subcommands to listen to PlanktoScope state updates via the MQTT API
- Added subcommand to start a segmentation routine and wait until it finishes (by default) or until it starts (by command-line flags) or not wait at all (by command-line flags)
