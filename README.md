# TCI Go Client Library

This is a client library for Expert Electornic's [TCI protocol](https://github.com/maksimus1210/TCI) written in Go. It comes with a simple CLI client application that allows to send single commands to the TCI server and to monitor incoming TCI messages.

Currently, there is no support of IQ or audio data.

## Some Details About TCI

Things that I did not find in the TCI reference document and that may be useful if you also want to develop a client:

* TCI is based on a websocket connection. Commands are send as text messages, IQ data is send as binary messages.
* When you send a command you need to wait until the server confirms it by returning your command in response, before you can send the next command.

## Build

This library does not have any fancy dependencies, so it can be build with a simple:

```
go build
```

## Install

To install the CLI client application, simply use the `go install` command:

```
go install github.com/ftl/tci
```

For more information about how to use the CLI client application, simply run the command `tci --help`.

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
