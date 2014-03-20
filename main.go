/*
	This file is part of LogHub.

	LogHub is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	LogHub is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with LogHub.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

const dateTimeFormat = "2006-01-02 15:04:05"

var commands = map[string]func([]string){
	"log":  logCommand,
	"hub":  hubCommand,
	"get":  getCommand,
	"put":  putCommand,
	"help": helpCommand,
}

func main() {
	if len(os.Args) < 2 {
		printCommands()
		os.Exit(1)
	}

	if cmd, found := commands[os.Args[1]]; found {
		cmd(os.Args[2:])
	}
}

func printCommands() {
	fmt.Println("Usage: loghub <command> <flags>")
	fmt.Println("Commands:")
	fmt.Println("  log  Starts log.")
	fmt.Println("  hub  Starts hub.")
	fmt.Println("  get  Gets log entries from log or hub.")
	fmt.Println("  put  Puts log entries to log or hub.")
	fmt.Println()
	fmt.Println("See 'loghub help <command>' for more information on a specific command.")
}

func logCommand(args []string) {
	flags := flag.NewFlagSet("log", flag.ExitOnError)
	address := flags.String("listen", ":10000", "Address and port to listen.")
	home := flags.String("home", "", "Home directory.") //TODO: Get default from environment variable.
	hub := flags.String("hub", "", "Hub address.")
	resistanceLevel := flags.Int64("resist", 1024, "Resistance level in megabytes.")
	statInerval := flags.Duration("stat", time.Second*10, "Status sending interval.")

	flags.Parse(args)

	if *home == "" {
		println("Home directory is not specified.")
		os.Exit(1)
	}

	if homeStat, err := os.Stat(*home); err != nil || !homeStat.IsDir() {
		println("Home doesn't exist or is not a directory.")
		os.Exit(1)
	}

	if *hub == "" {
		println("Hub is not specified.")
		os.Exit(1)
	}

	var port int

	if addr, err := net.ResolveTCPAddr("tcp", *address); err != nil {
		println("Failed to resolve the address:", err.Error(), ".")
		os.Exit(1)
	} else {
		port = addr.Port
	}

	logManager := NewDefaultLogManager(*home)

	var stopLogStatSender func()

	if s, err := startLogStatSender(*hub, logManager, port, *resistanceLevel, *statInerval); err != nil {
		println("Failed to start stat sender:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopLogStatSender = s
	}

	var stopServer func()

	if s, err := startServer(*address, NewLogMessageHandler(logManager)); err != nil {
		println("Failed to start server:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopServer = s
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)

	for _ = range signals {
		stopServer()
		stopLogStatSender()
		break
	}
}

func hubCommand(args []string) {

}

func getCommand(args []string) {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of log or hub.")
	baseStr := flags.String("base", "", "Base date and time.")
	rng := flags.Duration("range", 0, "Time range of timestamps relative to the base.")
	minSev := flags.Int("minsev", 0, "Min severity.")
	maxSev := flags.Int("maxsev", 0, "Max severity.")
	src := flags.String("src", "", "Log sources.")
	format := flags.String("fmt", "%s %s %d %s", "Log entry format.")
	tsfmt := flags.String("tsfmt", "2006-01-02 15:04:05", "Timestamp format.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	var base time.Time

	if *baseStr == "" {
		base = time.Now()
	} else {
		if t, err := time.Parse(dateTimeFormat, *baseStr); err != nil {
			println("Base date/time is not in correct format. Must be 'YYYY-MM-DD hh:mm:ss'.")
			os.Exit(1)
		} else {
			base = t
		}
	}

	var from, to int64

	if *rng < 0 {
		from = base.Add(*rng).UnixNano()
		to = base.UnixNano()
	} else {
		from = base.UnixNano()
		to = base.Add(*rng).UnixNano()
	}

	client := NewLogHubClient(*addr, 1)
	defer client.Close()

	queries := make(chan *LogQueryJSON)
	results := make(chan *OutgoingLogEntryJSON)

	client.Read(queries, results)

	if *src == "" {
		queries <- &LogQueryJSON{from, to, *minSev, *maxSev, *src}
	} else {
		for _, s := range strings.Split(*src, ",") {
			queries <- &LogQueryJSON{from, to, *minSev, *maxSev, s}
		}
	}

	close(queries)

	formatText := func(ent *OutgoingLogEntryJSON) {
		fmt.Printf(*format, time.Unix(0, ent.Ts).Format(*tsfmt), ent.Src, ent.Sev, ent.Msg)
		fmt.Println()
	}

	formatJSON := func(ent *OutgoingLogEntryJSON) {
		if bytes, err := json.Marshal(ent); err == nil {
			fmt.Println(string(bytes))
		}
	}

	var formatEntry func(*OutgoingLogEntryJSON)

	if *format == "JSON" {
		formatEntry = formatJSON
	} else {
		formatEntry = formatText
	}

	for ent := range results {
		formatEntry(ent)
	}
}

func putCommand(args []string) {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of log.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	input := bufio.NewReader(os.Stdin)
	client := NewLogHubClient(*addr, 1)
	entChan := make(chan *IncomingLogEntryJSON)

	client.Write(entChan)

	defer func() {
		close(entChan)
		client.Close()
	}()

	for {
		if line, _, err := input.ReadLine(); err == nil {
			if len(line) == 0 {
				break
			}

			ent := new(IncomingLogEntryJSON)

			if err = json.Unmarshal(line, ent); err == nil && ent.IsValid() {
				entChan <- ent
			}
		} else {
			break
		}
	}
}

func helpCommand(args []string) {

}
