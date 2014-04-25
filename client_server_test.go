// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/dbratus/loghub/lhproto"
	"github.com/dbratus/loghub/trace"
	"testing"
	"time"
)

func TestClientServer(t *testing.T) {
	//trace.SetTraceLevel(trace.LevelDebug)
	//defer trace.SetTraceLevel(trace.LevelError)

	cred := lhproto.Credentials{"", ""}
	serverAddress := "127.0.0.1:9999"
	messageHandler := newTestProtocolHandler()
	var closeServer func()

	if c, err := startServer(serverAddress, messageHandler, nil, ""); err != nil {
		t.Errorf("Failed to start LogHub server", err.Error())
		t.FailNow()
	} else {
		closeServer = c
	}

	client := lhproto.NewClient(serverAddress, 1, false, false)

	entriesToWrite := make(chan *lhproto.IncomingLogEntryJSON)
	client.Write(&cred, entriesToWrite)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &lhproto.IncomingLogEntryJSON{1, "Source", "Message"}

		entriesToWrite <- m
	}

	close(entriesToWrite)

	incomingMsgCnt := 0
	for incomingMsgCnt < testLogEntriesCount {
		select {
		case <-messageHandler.entriesWritten:
			incomingMsgCnt++

		case <-time.After(time.Second * 10):
			t.Errorf("Incomming entries have not arrived. Expected %d, got %d.", testLogEntriesCount, incomingMsgCnt)
			t.FailNow()
		}
	}

	truncateCmd := lhproto.TruncateJSON{"src", 10000}
	client.Truncate(&cred, &truncateCmd)

	select {
	case cmd := <-messageHandler.truncations:
		if cmd.Lim != truncateCmd.Lim {
			t.Error("Lim doesn't match.")
			t.FailNow()
		}

		if cmd.Src != truncateCmd.Src {
			t.Error("Src doesn't match.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("Truncation has not arrived.")
		t.FailNow()
	}

	transferCmd := lhproto.TransferJSON{1, ":10000", 1024}
	client.Transfer(&cred, &transferCmd)

	select {
	case cmd := <-messageHandler.transfers:
		if cmd.Lim != transferCmd.Lim {
			t.Error("Lim doesn't match.")
			t.FailNow()
		}

		if cmd.Addr != transferCmd.Addr {
			t.Error("Addr doesn't match.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("Transfer has not arrived.")
		t.FailNow()
	}

	acceptCmd := lhproto.AcceptJSON{"src/file", 1}
	acceptChan := make(chan *lhproto.InternalLogEntryJSON)
	acceptResult := make(chan *lhproto.AcceptResultJSON)

	client.Accept(&cred, &acceptCmd, acceptChan, acceptResult)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &lhproto.InternalLogEntryJSON{1, "src", EncodingPlain, "Message", timeToTimestamp(time.Now())}

		acceptChan <- m
	}

	close(acceptChan)

	select {
	case cmd := <-messageHandler.accepts:
		if cmd.Chunk != acceptCmd.Chunk {
			t.Error("Chunk doesn't match.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("Transfer has not arrived.")
		t.FailNow()
	}

	acceptedMsgCnt := 0
	for acceptedMsgCnt < testLogEntriesCount {
		select {
		case <-messageHandler.entriesAccepted:
			acceptedMsgCnt++

		case <-time.After(time.Second * 10):
			t.Errorf("Accepted entries have not arrived. Expected %d, got %d.", testLogEntriesCount, acceptedMsgCnt)
			t.FailNow()
		}
	}

	select {
	case cmd := <-acceptResult:
		if !cmd.Result {
			t.Error("Accept failed.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("Accept result has not arrived.")
		t.FailNow()
	}

	statResult := make(chan *lhproto.StatJSON)

	client.Stat(&cred, statResult)

	statCnt := 0
	for statCnt < testStatCnt {
		select {
		case <-statResult:
			statCnt++

		case <-time.After(time.Second * 10):
			t.Errorf("Stats have not arrived. Expected %d, got %d.", testStatCnt, statCnt)
			t.FailNow()
		}
	}

	userRoles := []string{"role1", "role2"}
	userCmd := lhproto.UserInfoJSON{"username", "password", true, userRoles, true}
	client.User(&cred, &userCmd)

	select {
	case cmd := <-messageHandler.users:
		if cmd.Name != userCmd.Name {
			t.Error("Name doesn't match.")
			t.FailNow()
		}

		if cmd.Password != userCmd.Password {
			t.Error("Password doesn't match.")
			t.FailNow()
		}

		if cmd.SetPassword != userCmd.SetPassword {
			t.Error("SetPassword doesn't match.")
			t.FailNow()
		}

		if len(cmd.Roles) != len(userCmd.Roles) || cmd.Roles[0] != userCmd.Roles[0] || cmd.Roles[1] != userCmd.Roles[1] {
			t.Error("Roles doesn't match.")
			t.FailNow()
		}

		if cmd.Delete != userCmd.Delete {
			t.Error("Delete doesn't match.")
			t.FailNow()
		}
	case <-time.After(time.Second * 10):
		t.Error("User has not arrived.")
		t.FailNow()
	}

	newPass := lhproto.PasswordJSON{"password"}
	client.Password(&cred, &newPass)

	select {
	case pass := <-messageHandler.passwords:
		if pass.Password != newPass.Password {
			t.Error("Password doesn't match.")
			t.FailNow()
		}

	case <-time.After(time.Second * 10):
		t.Error("User has not arrived.")
		t.FailNow()
	}

	queries := make(chan *lhproto.LogQueryJSON)
	result := make(chan *lhproto.OutgoingLogEntryJSON)

	client.Read(&cred, queries, result)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(lhproto.LogQueryJSON)
	}

	close(queries)

	resultLen := 0
	for _ = range result {
		resultLen++
	}

	if resultLen < testLogEntriesCount {
		t.Errorf(
			"Failed to read outgoing log entries through the JSON client. %d read, expected %d.",
			resultLen,
			testLogEntriesCount,
		)
		t.FailNow()
	}

	queries = make(chan *lhproto.LogQueryJSON)
	resultInternal := make(chan *lhproto.InternalLogEntryJSON)

	client.InternalRead(&cred, queries, resultInternal)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(lhproto.LogQueryJSON)
	}

	close(queries)

	resultLen = 0
	for _ = range resultInternal {
		resultLen++
	}

	if resultLen < testLogEntriesCount {
		t.Errorf(
			"Failed to read internal log entries through the JSON client. %d read, expected %d.",
			resultLen,
			testLogEntriesCount,
		)
		t.FailNow()
	}

	client.Close()
	closeServer()

	if !messageHandler.IsClosed() {
		t.Error("Failed to close JSON message handler.")
		t.FailNow()
	}
}

func TestClientWithoutServer(t *testing.T) {
	trace.SetTraceLevel(-1)
	defer trace.SetTraceLevel(trace.LevelError)

	client := lhproto.NewClient(":9999", 1, false, false)

	cred := lhproto.Credentials{"", ""}
	entriesToWrite := make(chan *lhproto.IncomingLogEntryJSON)
	client.Write(&cred, entriesToWrite)

	for i := 0; i < testLogEntriesCount; i++ {
		m := &lhproto.IncomingLogEntryJSON{1, "Source", "Message"}

		entriesToWrite <- m
	}

	close(entriesToWrite)

	queries := make(chan *lhproto.LogQueryJSON)
	result := make(chan *lhproto.OutgoingLogEntryJSON)

	client.Read(&cred, queries, result)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(lhproto.LogQueryJSON)
	}

	close(queries)

	lhproto.PurgeOutgoingLogEntryJSON(result)

	queries = make(chan *lhproto.LogQueryJSON)
	resultInternal := make(chan *lhproto.InternalLogEntryJSON)

	client.InternalRead(&cred, queries, resultInternal)

	for i := 0; i < testLogQueriesCount; i++ {
		queries <- new(lhproto.LogQueryJSON)
	}

	close(queries)

	lhproto.PurgeInternalLogEntryJSON(resultInternal)

	client.Close()
}
