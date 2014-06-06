// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dbratus/loghub/auth"
	"github.com/dbratus/loghub/lhproto"
	"github.com/dbratus/loghub/trace"
	"github.com/dbratus/loghub/webui"
	"github.com/howeyc/gopass"
	"os"
	"os/signal"
	"strings"
	"time"
)

const dateTimeFormat = "2006-01-02 15:04:05"

type consoleCommand struct {
	command     func([]string)
	description string
}

var commands = map[string]consoleCommand{
	"log":      consoleCommand{logCommand, "Starts log."},
	"hub":      consoleCommand{hubCommand, "Starts hub."},
	"ui":       consoleCommand{uiCommand, "Starts web UI."},
	"get":      consoleCommand{getCommand, "Gets log entries from log or hub."},
	"put":      consoleCommand{putCommand, "Puts log entries to log or hub."},
	"truncate": consoleCommand{truncateCommand, "Truncates the log."},
	"stat":     consoleCommand{statCommand, "Gets stats of a log or hub."},
	"user":     consoleCommand{userCommand, "Manages user accounts."},
	"pass":     consoleCommand{passCommand, "Changes user's password."},
	"help":     consoleCommand{helpCommand, "Prints help for a specific command."},
}

func main() {
	if len(os.Args) < 2 {
		printCommands()
		os.Exit(1)
	}

	if cmd, found := commands[os.Args[1]]; found {
		cmd.command(os.Args[2:])
	}
}

func printCommands() {
	fmt.Println("Usage: loghub <command> <flags>")
	fmt.Println("Commands:")

	maxCmdLen := 0

	for cmd, _ := range commands {
		l := len(cmd)

		if l > maxCmdLen {
			maxCmdLen = l
		}
	}

	for cmd, inf := range commands {
		fmt.Printf("  %s%s%s\n", cmd, strings.Repeat(" ", maxCmdLen-len(cmd)+1), inf.description)
	}

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
	certFile := flags.String("cert", "", "TLS certificate PEM file.")
	keyFile := flags.String("key", "", "Private key PEM file.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol to connect logs.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")
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

	var cert *tls.Certificate = nil

	if *certFile != "" && *keyFile != "" {
		if c, err := tls.LoadX509KeyPair(*certFile, *keyFile); err != nil {
			println("Failed to load CA certificate:", err.Error())
			os.Exit(1)
		} else {
			cert = &c
		}
	}

	logManager := NewDefaultLogManager(*home)

	var stopLogStatSender func()

	lastTransferId := new(int64)
	limBytes := *lim * 1024 * 1024

	if s, err := startLogStatSender(*hub, logManager, *address, limBytes, lastTransferId, *statInerval); err != nil {
		println("Failed to start the stat sender:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopLogStatSender = s
	}

	var stopServer func()

	if s, err := startServer(*address, NewLogProtocolHandler(logManager, lastTransferId, limBytes, *useTLS, *trust), cert, *home); err != nil {
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
	home := flags.String("home", "", "Home directory.") //TODO: Get default from environment variable.
	certFile := flags.String("cert", "", "TLS certificate PEM file.")
	keyFile := flags.String("key", "", "Private key PEM file.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol to connect logs.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")
	debug := flags.Bool("debug", false, "Write debug information.")

	flags.Parse(args)

	if *home == "" {
		println("Home directory is not specified.")
		os.Exit(1)
	}

	if *debug {
		trace.SetTraceLevel(trace.LevelDebug)
	}

	var cert *tls.Certificate = nil

	if *certFile != "" && *keyFile != "" {
		if c, err := tls.LoadX509KeyPair(*certFile, *keyFile); err != nil {
			println("Failed to load CA certificate:", err.Error())
			os.Exit(1)
		} else {
			cert = &c
		}
	}

	var instanceKey string

	if k, err := auth.LoadInstanceKey(*home); err != nil {
		println("Failed to load instance key:", err.Error())
		os.Exit(1)
	} else {
		instanceKey = k
	}

	hub := NewDefaultHub(*useTLS, *trust, instanceKey)

	var stopLogStatReceiver func()

	if s, err := startLogStatReceiver(*statAddress, hub); err != nil {
		println("Failed to start the stat receiver:", err.Error(), ".")
		os.Exit(1)
	} else {
		stopLogStatReceiver = s
	}

	var stopServer func()

	if s, err := startServer(*address, NewHubProtocolHandler(hub), cert, *home); err != nil {
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

func uiCommand(args []string) {
	flags := flag.NewFlagSet("ui", flag.ExitOnError)
	listenAddr := flags.String("http", ":8080", "Address and port to listen.")
	address := flags.String("addr", ":10000", "Address and port of a log or hub.")
	certFile := flags.String("cert", "", "TLS certificate PEM file.")
	keyFile := flags.String("key", "", "Private key PEM file.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol to connect logs.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")
	debug := flags.Bool("debug", false, "Write debug information.")

	flags.Parse(args)

	if *debug {
		trace.SetTraceLevel(trace.LevelDebug)
	}

	webui.Start(*listenAddr, *certFile, *keyFile, *address, *useTLS, *trust)
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
	user := flags.String("u", auth.Anonymous, "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	var base time.Time

	if *baseStr == "" {
		base = time.Now()
	} else {
		if t, err := time.ParseInLocation(dateTimeFormat, *baseStr, time.Local); err != nil {
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

	password := ""

	if *user != auth.Anonymous {
		fmt.Print("Enter password:")
		password = string(gopass.GetPasswd())
	}

	cred := lhproto.Credentials{*user, password}

	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	defer client.Close()

	queries := make(chan *lhproto.LogQueryJSON)
	results := make(chan *lhproto.OutgoingLogEntryJSON)

	client.Read(&cred, queries, results)

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
	user := flags.String("u", auth.Anonymous, "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	input := bufio.NewReader(os.Stdin)
	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	entChan := make(chan *lhproto.IncomingLogEntryJSON)

	password := ""

	if *user != auth.Anonymous {
		fmt.Print("Enter password:")
		password = string(gopass.GetPasswd())
	}

	cred := lhproto.Credentials{*user, password}

	client.Write(&cred, entChan)

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
	user := flags.String("u", auth.Anonymous, "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	if *lim == time.Duration(0) {
		println("The limit is not specified.")
		os.Exit(1)
	}

	password := ""

	if *user != auth.Anonymous {
		fmt.Print("Enter password:")
		password = string(gopass.GetPasswd())
	}

	cred := lhproto.Credentials{*user, password}

	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	defer client.Close()

	if *src == "" {
		client.Truncate(&cred, &lhproto.TruncateJSON{*src, timeToTimestamp(time.Now().Add(*lim))})
	} else {
		for _, s := range strings.Split(*src, ",") {
			client.Truncate(
				&cred,
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
	user := flags.String("u", auth.Anonymous, "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")

	flags.Parse(args)

	if *addr == "" {
		println("Log or hub address is not specified.")
		os.Exit(1)
	}

	password := ""

	if *user != auth.Anonymous {
		fmt.Print("Enter password:")
		password = string(gopass.GetPasswd())
	}

	cred := lhproto.Credentials{*user, password}

	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	defer client.Close()

	stats := make(chan *lhproto.StatJSON)

	client.Stat(&cred, stats)

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

	totalSize := int64(0)
	totalLim := int64(0)

	for stat := range stats {
		fmt.Printf("%s %s/%s %.2f%%\n", stat.Addr, formatSize(stat.Sz), formatSize(stat.Lim), float64(stat.Sz)*100.0/float64(stat.Lim))

		totalSize += stat.Sz
		totalLim += stat.Lim
	}

	percentFull := float64(0)

	if totalLim > 0 {
		percentFull = float64(totalSize) * 100.0 / float64(totalLim)
	}

	fmt.Printf("Total %s/%s %.2f%%\n", formatSize(totalSize), formatSize(totalLim), percentFull)
}

func userCommand(args []string) {
	flags := flag.NewFlagSet("user", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of a log or a hub.")
	user := flags.String("u", auth.Anonymous, "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")
	name := flags.String("name", "", "Name of the user to edit.")
	pass := flags.Bool("pass", false, "Whether to set the password.")
	roles := flags.String("roles", "", "Comma separated list of roles to assign to the user.")
	del := flags.Bool("d", false, "Delete user.")

	flags.Parse(args)

	if *name == "" {
		println("Name is not specified.")
		os.Exit(1)
	}

	password := ""

	if *user != auth.Anonymous {
		fmt.Print("Enter password:")
		password = string(gopass.GetPasswd())
	}

	userPassword := ""

	if *pass {
		fmt.Print("Enter user's password:")
		userPassword = string(gopass.GetPasswd())
	}

	var rolesToAssign []string

	if *roles != "" {
		rolesToAssign = make([]string, 0, 4)

		for _, r := range strings.Split(*roles, ",") {
			rolesToAssign = append(rolesToAssign, strings.Trim(r, " "))
		}
	}

	cred := lhproto.Credentials{*user, password}

	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	defer client.Close()

	usr := lhproto.UserInfoJSON{
		*name,
		userPassword,
		*pass,
		rolesToAssign,
		*del,
	}

	client.User(&cred, &usr)
}

func passCommand(args []string) {
	flags := flag.NewFlagSet("pass", flag.ExitOnError)
	addr := flags.String("addr", "", "Address and port of a log or a hub.")
	user := flags.String("u", "", "User name.")
	useTLS := flags.Bool("tls", false, "Whether to use TLS protocol.")
	trust := flags.Bool("trust", false, "Whether to trust any server certificate.")

	flags.Parse(args)

	if *user == "" {
		println("User is not specified.")
		os.Exit(1)
	}

	fmt.Print("Enter password:")
	password := string(gopass.GetPasswd())

	fmt.Print("Enter new password:")
	newPassword := string(gopass.GetPasswd())

	cred := lhproto.Credentials{*user, password}

	client := lhproto.NewClient(*addr, 1, *useTLS, *trust)
	defer client.Close()

	client.Password(&cred, &lhproto.PasswordJSON{newPassword})
}

func helpCommand(args []string) {
}
