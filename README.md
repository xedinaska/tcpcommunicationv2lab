# Description

App is simple TCP server & client. You can find runnable binaries for all platforms in `./bin` folder.
You can build your own app using commands in Makefile (both for client and server)

## Code style

- To fix code format using `gofmt` tool just run `make fmt` command;

## Built With

* Only Golang standard library is used.

## Server usage:

- Run binary using provided binary or `make run` command.

## Client usage:

- Run binary using provided binary or `make run` command.
- Send commands to server in `COMMAND:MESSAGE` format. Use `STOP:` to stop client. Use `SEND:MESSAGE TEXT HERE` to send message from client to server.


## Available commands (client):

- `{"type": "command", "payload": "CLIENTS_LIST"}` - use this to get list of connected clients
- `{"type": "command", "payload": "STOP"}` - use this to stop the client
- `{"type": "message", "payload": {"ip": "127.0.0.1:54648", "text": "hello from new client"}}` - use this to send message to another client by ip
- `{"type": "message", "payload": {"id": "3132372e302e302e313a3534353438d41d8cd98f00b204e9800998ecf8427e", "text": "hello!"}}` - use this to send message to another client by ID
