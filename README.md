# PlanktoScope/cli
Command-line tool to operate and manage PlanktoScopes

This tool provides a command-line interface to the APIs exposed by PlanktoScopes for local and remote operation.

## Usage

### Download/install

First, you will need to download the tool, which is available as a single self-contained executable file, specifically a binary named `planktoscope`. You should visit this repository's [releases page](https://github.com/PlanktoScope/cli/releases/latest) and download an archive file for your platform and CPU architecture; for example, on a Raspberry Pi 4, you should download the archive named `planktoscope-cli_{version number}_linux_arm.tar.gz` (where the version number should be substituted). You can extract the `planktoscope` binary from the archive using a command like:
```
tar -xzf planktoscope-cli_{version number}_{os}_{cpu architecture}.tar.gz planktoscope
```

Then you may need to move the `planktoscope` binary into a directory in your system path, or you can just run the `planktoscope` binary in your current directory (in which case you should replace `planktoscope` with `./planktoscope` in the commands listed below), or you can just run the `planktoscope` binary by its absolute/relative path (in which case you should replace `planktoscope` with the absolute/relative path of the binary in the commands listed below).

## Licensing

Except where otherwise indicated, source code provided here is covered by the following information:

Copyright Ethan Li and PlanktoScope project contributors

SPDX-License-Identifier: Apache-2.0 OR BlueOak-1.0.0

You can use the source code provided here either under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0) or under the [Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0); you get to decide. We are making the software available under the Apache license because it's [OSI-approved](https://writing.kemitchell.com/2019/05/05/Rely-on-OSI.html), but we like the Blue Oak Model License more because it's easier to read and understand.
