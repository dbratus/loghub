// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.package main

package main

import (
	"os"
	"github.com/dbratus/loghub/trace"
)

var tmpDirTrace = trace.New("TempDir")

func getTempDir(name string) string {
	tmpDir := os.TempDir()

	if tmpDir[len(tmpDir)-1:] != "/" {
		tmpDir = tmpDir + "/"
	}

	return tmpDir + name
}

func makeTempDir(path string) {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if e := os.Mkdir(path, 0777); e != nil {
				tmpDirTrace.Error(e.Error())
			}
		} else {
			tmpDirTrace.Error(err.Error())
		}
	} else if !stat.IsDir() {
		tmpDirTrace.Error(path + " already exists and its not a directory.")
	} else {
		if err := os.RemoveAll(path); err != nil {
			tmpDirTrace.Error(err.Error())
		}

		if e := os.Mkdir(path, 0777); e != nil {
			tmpDirTrace.Error(e.Error())
		}
	}
}

func rmTempDir(path string) {
	if err := os.RemoveAll(path); err != nil {
		tmpDirTrace.Error(err.Error())
	}
}