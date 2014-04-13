// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package auth

import (
	"github.com/dbratus/loghub/tmpdir"
	"strings"
	"testing"
)

var actionsAllowed = map[string]string{
	"reader": "read,pass",
	"writer": "write,pass",
	"hub":    "read,iread,transfer,accept,truncate,stat,user,pass",
	"admin":  "read,truncate,stat,user,pass",
}

var actionsDenied = map[string]string{
	"reader":  "write,iread,transfer,accept,truncate,stat,user",
	"writer":  "read,iread,transfer,accept,truncate,stat,user",
	"hub":     "write",
	"admin":   "write,iread,transfer,accept",
	Anonymous: "read,write,iread,transfer,accept,truncate,stat,user,pass",
}

func setupPermissions(perms *Permissions) {
	reader := [...]string{"reader"}
	writer := [...]string{"writer"}
	hub := [...]string{"hub"}
	admin := [...]string{"admin"}

	perms.SetPassword("reader", "reader_password")
	perms.SetRoles("reader", reader[:])

	perms.SetPassword("writer", "writer_password")
	perms.SetRoles("writer", writer[:])

	perms.SetPassword("hub", "hub_password")
	perms.SetRoles("hub", hub[:])

	perms.SetPassword("admin", "admin_password")
	perms.SetRoles("admin", admin[:])
}

func checkPermissions(perms *Permissions, user string, password string, isAllowed bool, permissions string, t *testing.T) {
	for _, action := range strings.Split(permissions, ",") {
		if perms.IsAllowed(action, user, password) != isAllowed {
			t.Errorf("Authorization failed: %s %s %s", user, password, action)
		}
	}
}

func TestAuthenticationAuthorization(t *testing.T) {
	perms := new(Permissions)

	setupPermissions(perms)

	checkPermissions(perms, "reader", "reader_password", true, actionsAllowed["reader"], t)
	checkPermissions(perms, "writer", "writer_password", true, actionsAllowed["writer"], t)
	checkPermissions(perms, "hub", "hub_password", true, actionsAllowed["hub"], t)
	checkPermissions(perms, "admin", "admin_password", true, actionsAllowed["admin"], t)

	checkPermissions(perms, "reader", "wrong_password", false, actionsAllowed["reader"], t)
	checkPermissions(perms, "writer", "wrong_password", false, actionsAllowed["writer"], t)
	checkPermissions(perms, "hub", "wrong_password", false, actionsAllowed["hub"], t)
	checkPermissions(perms, "admin", "wrong_password", false, actionsAllowed["admin"], t)

	checkPermissions(perms, "reader", "reader_password", false, actionsDenied["reader"], t)
	checkPermissions(perms, "writer", "writer_password", false, actionsDenied["writer"], t)
	checkPermissions(perms, "hub", "hub_password", false, actionsDenied["hub"], t)
	checkPermissions(perms, "admin", "admin_password", false, actionsDenied["admin"], t)
	checkPermissions(perms, Anonymous, "", false, actionsDenied[Anonymous], t)

	perms.DeleteUser("reader")
	perms.DeleteUser("writer")

	checkPermissions(perms, "reader", "reader_password", false, actionsAllowed["reader"], t)
	checkPermissions(perms, "writer", "writer_password", false, actionsAllowed["writer"], t)

	anonPerms := [...]string{"writer"}
	perms.SetRoles(Anonymous, anonPerms[:])

	checkPermissions(perms, Anonymous, "", true, "write", t)
}

func TestLoadSave(t *testing.T) {
	home := tmpdir.GetPath("auth.test")

	tmpdir.Make(home)
	defer tmpdir.Rm(home)

	perms := new(Permissions)

	setupPermissions(perms)

	if err := perms.Save(home); err != nil {
		t.Errorf("Failed to save permissions: %s.", err.Error())
		t.FailNow()
	}

	if p, err := LoadPermissions(home); err != nil {
		t.Errorf("Failed to load permissions: %s.", err.Error())
		t.FailNow()
	} else {
		perms = p
	}

	checkPermissions(perms, "reader", "reader_password", true, actionsAllowed["reader"], t)
	checkPermissions(perms, "writer", "writer_password", true, actionsAllowed["writer"], t)
	checkPermissions(perms, "hub", "hub_password", true, actionsAllowed["hub"], t)
	checkPermissions(perms, "admin", "admin_password", true, actionsAllowed["admin"], t)
}
