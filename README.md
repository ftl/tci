# TCI Go Client Library

This is a client library for Expert Electronic's [TCI protocol](https://github.com/maksimus1210/TCI) written in Go. The currently supported version of TCI is 1.4.

The library comes with a simple CLI client application that allows to some simple things and is mainly ment as example on how to use this library:

* monitor incoming TCI message
* send text as CW
* generate and transmit a one- or two-tone-signal for simple measurements

## Some Details About TCI

Things that I did not find in the TCI reference document and that may be useful if you also want to develop a client:

* TCI is based on a websocket connection. Commands are send as text messages, IQ data is send as binary messages.
* When you send a command you need to wait until the server confirms it by returning your command in response, before you can send the next command.
* The binary messages come in little-endian byte order.
* Audio data always comes in stereo, even the tx audio.

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

## Disclaimer

I am in no association with [Expert Electronics](https://eesdr.com) of any form. I develop this library on my own, in my free time, for my own leisure. I hope the
time invested will also help others. I am not liable for anything that happens to you or your equipment through the use of this software. That said - have fun!

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
