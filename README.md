# JSON-RPC Mock server

Mock server for JSON-RPC with configuration in json file.

## How to build

Install [Golang](https://golang.org/) on your PC.

Download this project.

Open the project folder in the terminal.

Run `go build -o jrpc_mock main.go`.

Now you can see the program `jrpc_mock` in the project folder.

## Configuration

Move the file `jrpc_mock` and copy the file`jrpc-config.json` from the project directory.

Rename `jrpc-config.json` (example:` YOUR_OWN_PROJECT_NAME-VERSION-jrpc-config.json`).

Open the configuration file and configure it.

## Running

Open a terminal in the folder where you moved the program and config file.

Run `./jrpc_mock --config=YOUR_CONFIG_FILE`.
