// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"github.com/dbratus/loghub/lhproto"
	"os"
)

const (
	AllowRead = 1 << iota
	AllowWrite
	AllowInternalRead
	AllowTransfer
	AllowAccept
	AllowTruncate
	AllowStat
	AllowUser
	AllowPassword
)

const (
	RoleReader = 1 << iota
	RoleWriter
	RoleHub
	RoleAdmin
)

const (
	DefaultAdmin = "admin"
	DefaultHub   = "hub"
	Anonimous    = "all"
)

var rolePermissions = map[int]int{
	RoleReader: AllowRead | AllowPassword,
	RoleWriter: AllowWrite | AllowPassword,
	RoleAdmin:  AllowRead | AllowTruncate | AllowStat | AllowUser | AllowPassword,
	RoleHub:    AllowRead | AllowInternalRead | AllowTransfer | AllowAccept | AllowTruncate | AllowStat | AllowUser | AllowPassword,
}

var roleNames = map[string]int{
	"reader": RoleReader,
	"writer": RoleWriter,
	"hub":    RoleHub,
	"admin":  RoleAdmin,
}

var permissionNames = map[string]int{
	lhproto.ActionRead:         AllowRead,
	lhproto.ActionWrite:        AllowWrite,
	lhproto.ActionInternalRead: AllowInternalRead,
	lhproto.ActionTransfer:     AllowTransfer,
	lhproto.ActionAccept:       AllowAccept,
	lhproto.ActionTruncate:     AllowTruncate,
	lhproto.ActionStat:         AllowStat,
	lhproto.ActionUser:         AllowUser,
	lhproto.ActionPassword:     AllowPassword,
}

type Permissions struct {
	Users map[string]*UserData
}

type UserData struct {
	Roles        int
	PasswordHash []byte
}

func LoadPermissions(home string) (*Permissions, error) {
	backupFileName := home + "/permissions.bak"
	fileName := home + "/permissions"

	var perms Permissions

	//If backup file exists, this means that the last saving
	//has not been completed successfully, so the backuped permissions
	//need to be restored.
	if _, err := os.Stat(backupFileName); err == nil {
		if f, err := os.OpenFile(backupFileName, os.O_RDONLY, 0); err != nil {
			return nil, err
		} else {
			defer f.Close()

			if err := gob.NewDecoder(f).Decode(&perms); err != nil {
				return nil, err
			}

			if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
				return nil, err
			}

			if err := os.Rename(backupFileName, fileName); err != nil {
				return nil, err
			}
		}
	} else {
		if f, err := os.OpenFile(fileName, os.O_RDONLY, 0); err != nil {
			//If the permissions file doesn't exist,
			//creating default permissions.
			if os.IsNotExist(err) {
				adminRole := [...]string{"admin"}
				perms.SetRoles(DefaultAdmin, adminRole[:])
				perms.SetPassword(DefaultAdmin, "admin")

				hubRole := [...]string{"hub"}
				perms.SetRoles(DefaultHub, hubRole[:])
				perms.SetPassword(DefaultHub, "hub")

				perms.SetPassword(Anonimous, "")
			}
		} else {
			defer f.Close()

			if err := gob.NewDecoder(f).Decode(&perms); err != nil {
				return nil, err
			}
		}
	}

	return &perms, nil
}

func (perms *Permissions) checkInit() {
	if perms.Users == nil {
		perms.Users = make(map[string]*UserData)
	}
}

func (perms *Permissions) IsAllowed(action, user, password string) bool {
	perms.checkInit()

	if ud, found := perms.Users[user]; found {
		if password == "" && ud.PasswordHash == nil {
			return true
		}

		passwordHash := sha1.Sum([]byte(password))

		if bytes.Compare(passwordHash[:], ud.PasswordHash) != 0 {
			return false
		}

		userPermissions := 0

		for role, permissions := range rolePermissions {
			if ud.Roles&role != 0 {
				userPermissions |= permissions
			}
		}

		if perm, found := permissionNames[action]; found {
			return userPermissions&perm != 0
		} else {
			return true
		}
	} else {
		return false
	}
}

func (perms *Permissions) SetRoles(user string, roles []string) {
	perms.checkInit()

	userRoles := 0

	for _, roleName := range roles {
		if role, found := roleNames[roleName]; found {
			userRoles |= role
		}
	}

	if ud, found := perms.Users[user]; found {
		ud.Roles = userRoles
	} else {
		perms.Users[user] = &UserData{
			userRoles,
			nil,
		}
	}
}

func (perms *Permissions) SetPassword(user string, password string) {
	perms.checkInit()

	var passwordHash []byte

	if password != "" {
		h := sha1.Sum([]byte(password))
		passwordHash = h[:]
	}

	if ud, found := perms.Users[user]; found {
		ud.PasswordHash = passwordHash
	} else {
		perms.Users[user] = &UserData{
			0,
			passwordHash,
		}
	}
}

func (perms *Permissions) DeleteUser(user string) {
	if perms.Users != nil && user != DefaultAdmin && user != DefaultHub && user != Anonimous {
		delete(perms.Users, user)
	}
}

func (perms *Permissions) Save(home string) error {
	perms.checkInit()

	fileName := home + "/permissions"

	if f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0660); err != nil {
		return err
	} else {
		defer f.Close()

		if err := f.Truncate(0); err != nil {
			return err
		}

		if err := gob.NewEncoder(f).Encode(perms); err != nil {
			return err
		}
	}

	return nil
}
