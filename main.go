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
	home := flags.String("home", "", "Home directory.")

	flags.Parse(args)

	if *home == "" {
		println("Home directory is not specified.")
		os.Exit(1)
	}

	if homeStat, err := os.Stat(*home); err != nil || !homeStat.IsDir() {
		println("Home doesn't exist or is not a directory.")
		os.Exit(1)
	}

	var stopServer func()

	if s, err := startServer(*address, NewLogMessageHandler(NewDefaultLogManager(*home))); err != nil {
		println("Failed to start server:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopServer = s
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)

	for _ = range signals {
		stopServer()
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

	for ent := range results {
		//TODO: Accept format as argument.
		fmt.Printf("%s %d %s %s", time.Unix(0, ent.Ts).String(), ent.Sev, ent.Src, ent.Msg)
		fmt.Println()
	}
}

func putCommand(args []string) {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of log or hub.")

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
