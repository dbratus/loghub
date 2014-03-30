// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dbratus/loghub/lhproto"
	"github.com/dbratus/loghub/trace"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

const dateTimeFormat = "2006-01-02 15:04:05"

var commands = map[string]func([]string){
	"log":      logCommand,
	"hub":      hubCommand,
	"get":      getCommand,
	"put":      putCommand,
	"truncate": truncateCommand,
	"stat":     statCommand,
	"help":     helpCommand,
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
	fmt.Println("  log       Starts log.")
	fmt.Println("  hub       Starts hub.")
	fmt.Println("  get       Gets log entries from log or hub.")
	fmt.Println("  put       Puts log entries to log or hub.")
	fmt.Println("  truncate  Truncates the log.")
	fmt.Println("  stat      Gets stats of a log or hub.")
	fmt.Println()
	fmt.Println("See 'loghub help <command>' for more information on a specific command.")
}

func logCommand(args []string) {
	flags := flag.NewFlagSet("log", flag.ExitOnError)
	address := flags.String("listen", ":10000", "Address and port to listen.")
	home := flags.String("home", "", "Home directory.") //TODO: Get default from environment variable.
	hub := flags.String("hub", "", "Hub address.")
	lim := flags.Int64("lim", 1024, "Log size limit in megabytes.")
	statInerval := flags.Duration("stat", time.Second*10, "Status sending interval.")
	debug := flags.Bool("debug", false, "Write debug information.")

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

	if *debug {
		trace.SetTraceLevel(trace.LevelDebug)
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

	lastTransferId := new(int64)
	limBytes := *lim * 1024 * 1024

	if s, err := startLogStatSender(*hub, logManager, port, limBytes, lastTransferId, *statInerval); err != nil {
		println("Failed to start the stat sender:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopLogStatSender = s
	}

	var stopServer func()

	if s, err := startServer(*address, NewLogProtocolHandler(logManager, lastTransferId, limBytes)); err != nil {
		println("Failed to start the server:", err.Error(), ".")
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
	flags := flag.NewFlagSet("hub", flag.ExitOnError)
	statAddress := flags.String("stat", ":9999", "Address and port to collect stat.")
	address := flags.String("listen", ":10000", "Address and port to listen.")
	debug := flags.Bool("debug", false, "Write debug information.")

	flags.Parse(args)

	if *debug {
		trace.SetTraceLevel(trace.LevelDebug)
	}

	hub := NewDefaultHub()

	var stopLogStatReceiver func()

	if s, err := startLogStatReceiver(*statAddress, hub); err != nil {
		println("Failed to start the stat receiver:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopLogStatReceiver = s
	}

	var stopServer func()

	if s, err := startServer(*address, NewHubProtocolHandler(hub)); err != nil {
		println("Failed to start the server:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopServer = s
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)

	for _ = range signals {
		stopServer()
		stopLogStatReceiver()
		break
	}
}

func getCommand(args []string) {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of log or hub.")
	baseStr := flags.String("base", "", "Base date and time.")
	rng := flags.Duration("range", 0, "Time range of timestamps relative to the base.")
	minSev := flags.Int("minsev", 0, "Min severity.")
	maxSev := flags.Int("maxsev", 0, "Max severity.")
	src := flags.String("src", "", "Comma separated list of log sources.")
	format := flags.String("fmt", "%s %s %d %s", "Log entry format.")
	tsfmt := flags.String("tsfmt", "2006-01-02 15:04:05", "Timestamp format.")
	isUtc := flags.Bool("utc", false, "Return UTC timestamps.")

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
		from = timeToTimestamp(base.Add(*rng))
		to = timeToTimestamp(base)
	} else {
		from = timeToTimestamp(base)
		to = timeToTimestamp(base.Add(*rng))
	}

	client := lhproto.NewClient(*addr, 1)
	defer client.Close()

	queries := make(chan *lhproto.LogQueryJSON)
	results := make(chan *lhproto.OutgoingLogEntryJSON)

	client.Read(queries, results)

	if *src == "" {
		queries <- &lhproto.LogQueryJSON{from, to, *minSev, *maxSev, *src}
	} else {
		for _, s := range strings.Split(*src, ",") {
			queries <- &lhproto.LogQueryJSON{from, to, *minSev, *maxSev, strings.Trim(s, " ")}
		}
	}

	close(queries)

	formatText := func(ent *lhproto.OutgoingLogEntryJSON) {
		var tsToTime func(int64) time.Time

		if *isUtc {
			tsToTime = timestampToTime
		} else {
			tsToTime = timestampToLocalTime
		}

		fmt.Printf(*format, tsToTime(ent.Ts).Format(*tsfmt), ent.Src, ent.Sev, ent.Msg)
		fmt.Println()
	}

	formatJSON := func(ent *lhproto.OutgoingLogEntryJSON) {
		if bytes, err := json.Marshal(ent); err == nil {
			fmt.Println(string(bytes))
		}
	}

	var formatEntry func(*lhproto.OutgoingLogEntryJSON)

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
	flags := flag.NewFlagSet("put", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of log.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	input := bufio.NewReader(os.Stdin)
	client := lhproto.NewClient(*addr, 1)
	entChan := make(chan *lhproto.IncomingLogEntryJSON)

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

			ent := new(lhproto.IncomingLogEntryJSON)

			if err = json.Unmarshal(line, ent); err == nil && ent.IsValid() {
				entChan <- ent
			}
		} else {
			break
		}
	}
}

func truncateCommand(args []string) {
	flags := flag.NewFlagSet("truncate", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of a log or a hub.")
	src := flags.String("src", "", "Comma separated list of log sources.")
	lim := flags.Duration("lim", 0, "The limit of the truncation.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	if *lim == time.Duration(0) {
		println("The limit is not specified.")
		os.Exit(1)
	}

	client := lhproto.NewClient(*addr, 1)
	defer client.Close()

	if *src == "" {
		client.Truncate(&lhproto.TruncateJSON{*src, timeToTimestamp(time.Now().Add(*lim))})
	} else {
		for _, s := range strings.Split(*src, ",") {
			client.Truncate(
				&lhproto.TruncateJSON{
					strings.Trim(s, " "),
					timeToTimestamp(time.Now().Add(*lim)),
				},
			)
		}
	}
}

func statCommand(args []string) {
	flags := flag.NewFlagSet("stat", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of a log or a hub.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	client := lhproto.NewClient(*addr, 1)
	defer client.Close()

	stats := make(chan *lhproto.StatJSON)

	client.Stat(stats)

	formatSize := func(sz int64) string {
		kb := int64(1024)
		mb := kb * int64(1024)

		if sz < kb {
			return fmt.Sprintf("%db", sz)
		} else if sz < mb {
			return fmt.Sprintf("%dKb", sz/kb)
		} else {
			return fmt.Sprintf("%dMb", sz/mb)
		}
	}

	for stat := range stats {
		fmt.Printf("%s %s/%s %.2f%%\n", stat.Addr, formatSize(stat.Sz), formatSize(stat.Lim), float64(stat.Sz)*100.0/float64(stat.Lim))
	}
}

func helpCommand(args []string) {

}
