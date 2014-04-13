# LogHub

Large distributed systems may produce a lot of logs during integration testing. Services may need to write input and output XML messages for each call with a lot of intermediate data. In some cases the logs may grow with a speed of few gigabytes per hour, so managing of these logs becomes an issue. LogHub is intended to simplify querying and management of the large distributed logs by incorporating logging agents at multiple hosts into a single self-balancing storage with a single endpoint for querying.

LogHub runs as a daemon and can work in two modes - log and hub.

In the log mode, LogHub exposes TCP endpoint for writing and reading the log. The data is stored locally in the directory called LogHub home. The log is bound to a hub and provides it with the information on the log's configuration and state so that the hub can discover and connect to the log.

In the hub mode, LogHub serves two purposes - it provides a single access endpoint for multiple logs and it coordinates data transfers between them. Each log has a limit on its size. When the limit is exceeded, the log tries to push the data to other logs. However, the log itself knows nothing about other logs, so the hub tells the logs where they may push the data.

## Installation

To build LogHub from sources, you need Go infrastructure to be setup, $GOPATH/bin must be in PATH.

To get, build and install LogHub, run:

```
go get github.com/dbratus/loghub
```

## Starting LogHub

To start log, run LogHub with 'log' command:

```
loghub log -listen 127.0.0.1:10000 -home /var/loghub -hub 127.0.0.1:9999 -lim 10240
```

This command starts a 10Gb log, accepting connections at port 10000, storing the data at '/var/loghub' and sending notifications to the hub at 127.0.0.1:9999.

Loopback addresses may be omitted, so the shorter form of the command will be:

```
loghub log -listen :10000 -home /var/loghub -hub :9999 -lim 10240
```

To start hub, run LogHub with 'hub' command:

```
loghub hub -listen :10000 -stat :9999
```

This command starts a hub listening port 10000 for querying and accepting notifications from logs on port 9999.

## Security

For user authentication LogHub attaches passwords to messages, so, if you care about password security, it is important to use TLS protocol for communication between LogHub components.

To work in TLS mode, log and hub need a certificate and a private key in PEM format. Having these files, you can specify them as parameters to LogHub:

```
loghub log -listen 127.0.0.1:10000 -home /var/loghub -hub 127.0.0.1:9999 -lim 10240 -cert cert.pem -key privkey.pem
```

The 'hub' command accepts the same 'cert' and 'key' parameters, but additionally it needs 'tls' flag to know that the logs also work in TLS mode.

```
loghub hub -listen :10000 -stat :9999 -tls -cert cert.pem -key privkey.pem
```

### Permissions

As log or hub started for the first time, its not yet ready for use because nobody can read and write it. By default, LogHub creates the following users:

* admin - the default administrator.
* hub - the special user that may work as a hub.
* all - the default anonymous user.

The 'admin' and 'hub' users are created with default passwords 'admin' and 'hub' respectively, so the first thing you need to do creating a new log or hub is to set passwords for these default users.

```
loghub -addr :10000 -u admin -pass -name admin
loghub -addr :10000 -u admin -pass -name hub
```

The 'u' parameter is the user you are working as, 'name' specifies the name of the user you are editing, 'pass' flag tells that the password of the user needs to be changed. If the 'pass' flag is specified, you will be prompted for the user's password.

It is important to set hub's password every time you restart the hub. Due to security concerns, hub doesn't store its password persistently, so you need to tell it the password every time you start it.

To allow somebody to read and write the logs, you need to setup reader and writer users. This can be done with the same command.

```
loghub -addr :10000 -u admin -pass -name reader -roles reader
loghub -addr :10001 -u admin -pass -name writer -roles writer
loghub -addr :10002 -u admin -pass -name writer -roles writer
...
```

Readers may be setup only on hub if they are going to read the logs via the hub only. Writers, as they write directly to the logs, need to be setup for each log individually. In test environments, you may wish to allow anonymous users to read and write the logs. This can be done with the following command:

```
loghub -addr :10000 -u admin -name all -roles reader,writer
loghub -addr :10001 -u admin -name all -roles reader,writer
loghub -addr :10002 -u admin -name all -roles reader,writer
...
```

## Writing logs

Log entries in LogHub are the records with the following fields:

* Timestamp assigned by LogHub.
* Severity - an integer value within [0-255] indicating the severity of the logged event.
* Source - a string identifying the source of the event.
* Message - a string describing the event.

You can write LogHub logs by using client for your platform:

* [Go](https://github.com/dbratus/loghub-go).
* [Node.js](https://github.com/dbratus/loghub-js).

#### Go

```Go
package main

import "github.com/dbratus/loghub-go"

func main() {
	log := loghub.NewClient(":10001", 1)
	defer log.Close()

	log.Write(1, "Example", "Example message.")
}
```

### Node.js

```js
var http = require('http'),
	loghub = require('loghub');

//Connecting to log at localhost:10001.
var log = loghub.connect(10001);

var srv = http.createServer(function (req, resp) {
	//Writing log entry:
	//Severity = 1,
	//Source = 'HTTP',
	//Message = 'Got request.'.
	log.write(1, 'HTTP', 'Got request.');

	resp.write('OK');
	resp.end();
});

srv.listen(8080);
```

## Obtaining logs

You can obtain LogHub logs by using 'get' command:

```
loghub get -addr hostname:10000 -base "2014-01-15 10:00:00" -range 8h -minsev 0 -maxsev 10 -src "Source1, Source2"
```

The 'addr' parameter specifies the host and port of a log or hub to get the data from.

The 'base' and 'range' parameters specify the timespan of the log entries. The 'base' is a date and time in 'YYYY-MM-DD hh:mm:ss' format. It may be omitted, then the current time is taken as a base. The 'range' specifies timespan relative to the base. It may be negative that is useful when the base is the current time. For example, by specifying `-range -24h` you can get logs of the last 24 hours. The value of the 'range' parameter must be a valid "duration string". As defined in Go documentation: "A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".".

The 'minsev' and 'maxsev' specifies the range of severities.

The 'src' parameter is a comma separated list of regular expressions that the log sources must match. If this parameter is omitted, entries of all sources are returned.

The time of the timestamps by default is a local time. You can add `-utc` parameter to get them in UTC.

To get log entries in JSON format, add `-fmt JSON`. In this case the timestamps will be the number of nanoseconds since Unix epoch UTC.
