// Copyright (C) 2014 Dmitry Bratus
//
// The use of this source code is governed by the license
// that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
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
	RoleInstance
)

const (
	DefaultAdmin = "admin"
	Anonymous    = "all"
)

const instanceKeyLength = 32

var rolePermissions = map[int]int{
	RoleReader:   AllowRead | AllowPassword,
	RoleWriter:   AllowWrite | AllowPassword,
	RoleAdmin:    AllowRead | AllowTruncate | AllowStat | AllowUser | AllowPassword,
	RoleInstance: AllowWrite | AllowRead | AllowInternalRead | AllowTransfer | AllowAccept | AllowTruncate | AllowStat | AllowUser | AllowPassword,
}

var roleNames = map[string]int{
	"reader":   RoleReader,
	"writer":   RoleWriter,
	"admin":    RoleAdmin,
	"instance": RoleInstance,
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
	Users       map[string]*UserData
	instanceKey string
}

type UserData struct {
	Roles        int
	PasswordHash []byte
}

func LoadInstanceKey(home string) (string, error) {
	fileName := home + "/instance.key"

	if f, err := os.OpenFile(fileName, os.O_RDONLY, 0); err != nil {
		if os.IsNotExist(err) {
			key := make([]byte, instanceKeyLength)

			if _, err := rand.Read(key); err != nil {
				return "", err
			}

			keyBase64 := base64.StdEncoding.EncodeToString(key)

			if f, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0660); err != nil {
				return "", err
			} else {
				defer f.Close()

				if _, err := f.Write([]byte(keyBase64)); err != nil {
					return "", err
				}
			}

			return keyBase64, nil

		} else {
			return "", err
		}
	} else {
		defer f.Close()

		var key []byte

		if stat, err := f.Stat(); err != nil {
			return "", err
		} else {
			key = make([]byte, stat.Size())
		}

		if _, err := f.Read(key); err != nil {
			return "", err
		}

		return string(key), nil
	}
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
				perms.SetRoles(DefaultAdmin, []string{"admin"})
				perms.SetPassword(DefaultAdmin, "admin")

				perms.SetPassword(Anonymous, "")
			}
		} else {
			defer f.Close()

			if err := gob.NewDecoder(f).Decode(&perms); err != nil {
				return nil, err
			}
		}
	}

	if instKey, err := LoadInstanceKey(home); err != nil {
		return nil, err
	} else {
		perms.SetRoles(instKey, []string{"instance"})
		perms.SetPassword(instKey, "")

		perms.instanceKey = instKey
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
		if password != "" || ud.PasswordHash != nil {
			passwordHash := sha1.Sum([]byte(password))

			if bytes.Compare(passwordHash[:], ud.PasswordHash) != 0 {
				return false
			}
		}

		//It is always prohibited for an anonymous to
		//set password for the anonymous account.
		if user == Anonymous && action == "pass" {
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
	if perms.Users != nil && user != DefaultAdmin && user != Anonymous {
		delete(perms.Users, user)
	}
}

func (perms *Permissions) Save(home string) error {
	perms.checkInit()

	defer func() {
		perms.SetRoles(perms.instanceKey, []string{"instance"})
		perms.SetPassword(perms.instanceKey, "")
	}()

	perms.DeleteUser(perms.instanceKey)

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
