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
