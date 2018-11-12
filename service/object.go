// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

type Object interface {
	/**********
	 * Require obj implementation
	 **********/

	Save(isLocked bool) error
	Delete(isLocked bool) error

	NewEmptyObj() Object
	GetBaseObject() *BaseObject

	GetByID(isLocked bool) error
	GetKey(id *types.PttID, isLocked bool) ([]byte, error)
	GetNewObjByID(id *types.PttID, isLocked bool) (Object, error)
	Unmarshal(theBytes []byte) error

	SetUpdateTS(ts types.Timestamp)
	GetUpdateTS() types.Timestamp

	RemoveMeta()

	// data
	GetBlockInfo() BlockInfo
	SetBlockInfo(blockInfo BlockInfo) error

	// sync-info
	GetSyncInfo() SyncInfo
	SetSyncInfo(syncInfo SyncInfo) error

	/**********
	 * implemented in BaseObject
	 **********/

	SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap, entityID *types.PttID, fullDBPrefix []byte, fullDBIdxPrefix []byte)
	Lock() error
	Unlock() error
	RLock() error
	RUnlock() error

	SetVersion(v types.Version)

	SetCreateTS(ts types.Timestamp)

	SetCreatorID(id *types.PttID)
	GetCreatorID() *types.PttID

	SetUpdaterID(id *types.PttID)

	SetID(id *types.PttID)
	GetID() *types.PttID

	SetLogID(id *types.PttID)
	GetLogID() *types.PttID

	SetUpdateLogID(id *types.PttID)
	GetUpdateLogID() *types.PttID

	SetStatus(status types.Status)
	GetStatus() types.Status

	SetEntityID(id *types.PttID)
	GetEntityID() *types.PttID
}

type BaseObject struct {
	V         types.Version
	ID        *types.PttID
	CreateTS  types.Timestamp `json:"CT"`
	CreatorID *types.PttID    `json:"CID"`
	UpdaterID *types.PttID    `json:"UID"`
	EntityID  *types.PttID    `json:"e,omitempty"`

	LogID       *types.PttID `json:"l,omitempty"`
	UpdateLogID *types.PttID `json:"u,omitempty"`

	Status types.Status `json:"S"`

	db              *pttdb.LDBBatch
	dbLock          *types.LockMap
	fullDBPrefix    []byte
	fullDBIdxPrefix []byte
}

func NewObject(
	id *types.PttID,
	createTS types.Timestamp,
	creatorID *types.PttID,
	entityID *types.PttID,

	logID *types.PttID,

	status types.Status,

	db *pttdb.LDBBatch,
	dbLock *types.LockMap,

	fullDBPrefix []byte,

	fullDBIdxPrefix []byte,
) *BaseObject {

	return &BaseObject{
		V:         types.CurrentVersion,
		ID:        id,
		CreateTS:  createTS,
		CreatorID: creatorID,
		UpdaterID: creatorID,
		EntityID:  entityID,

		LogID: logID,

		Status: status,

		db:              db,
		dbLock:          dbLock,
		fullDBPrefix:    fullDBPrefix,
		fullDBIdxPrefix: fullDBIdxPrefix,
	}
}

func (o *BaseObject) SetDB(db *pttdb.LDBBatch, dbLock *types.LockMap, entityID *types.PttID, fullDBPrefix []byte, fullDBIdxPrefix []byte) {
	o.db = db
	o.dbLock = dbLock
	o.EntityID = entityID
	o.fullDBPrefix = fullDBPrefix
	o.fullDBIdxPrefix = fullDBIdxPrefix
}

func (o *BaseObject) Lock() error {
	return o.dbLock.Lock(o.ID)
}

func (o *BaseObject) Unlock() error {
	return o.dbLock.Unlock(o.ID)
}

func (o *BaseObject) RLock() error {
	return o.dbLock.Lock(o.ID)
}

func (o *BaseObject) RUnlock() error {
	return o.dbLock.Unlock(o.ID)
}

func (o *BaseObject) SetVersion(v types.Version) {
	o.V = v
}

func (o *BaseObject) SetCreateTS(ts types.Timestamp) {
	o.CreateTS = ts
}

func (o *BaseObject) SetCreatorID(id *types.PttID) {
	o.CreatorID = id
}

func (o *BaseObject) GetCreatorID() *types.PttID {
	return o.CreatorID
}

func (o *BaseObject) SetUpdaterID(id *types.PttID) {
	o.UpdaterID = id
}

func (o *BaseObject) SetID(id *types.PttID) {
	o.ID = id
}

func (o *BaseObject) GetID() *types.PttID {
	return o.ID
}

func (o *BaseObject) SetLogID(id *types.PttID) {
	o.LogID = id
}

func (o *BaseObject) GetLogID() *types.PttID {
	return o.LogID
}

func (o *BaseObject) SetUpdateLogID(id *types.PttID) {
	o.UpdateLogID = id
}

func (o *BaseObject) GetUpdateLogID() *types.PttID {
	return o.UpdateLogID
}

func (o *BaseObject) SetStatus(status types.Status) {
	o.Status = status
}

func (o *BaseObject) GetStatus() types.Status {
	return o.Status
}

func (o *BaseObject) SetEntityID(id *types.PttID) {
	o.EntityID = id
}

func (o *BaseObject) GetEntityID() *types.PttID {
	return o.EntityID
}

func (o *BaseObject) IdxPrefix() []byte {
	return o.fullDBIdxPrefix
}

func (o *BaseObject) IdxKey() ([]byte, error) {
	return append(o.fullDBIdxPrefix, o.ID[:]...), nil
}

func (o *BaseObject) Delete(isLocked bool) error {
	var err error

	log.Debug("Delete: start")

	if !isLocked {
		err = o.Lock()
		if err != nil {
			return err
		}
		defer o.Unlock()
	}

	idxKey, err := o.IdxKey()
	if err != nil {
		return err
	}

	err = o.db.DeleteAll(idxKey)
	if err != nil {
		return err
	}

	return nil
}

func (o *BaseObject) GetValueByID(isLocked bool) ([]byte, error) {
	var err error

	if !isLocked {
		err = o.RLock()
		if err != nil {
			return nil, err
		}
		defer o.RUnlock()
	}

	idxKey, err := o.IdxKey()
	if err != nil {
		return nil, err
	}

	log.Debug("GetValueByID: to GetByIdxKey", "idxKey", idxKey)
	val, err := o.db.GetByIdxKey(idxKey, 0)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (o *BaseObject) GetKey(id *types.PttID, isLocked bool) ([]byte, error) {
	if !isLocked {
		err := o.dbLock.RLock(id)
		if err != nil {
			return nil, err
		}
		defer o.dbLock.RUnlock(id)
	}

	o.ID = id
	idxKey, err := o.IdxKey()
	if err != nil {
		return nil, err
	}

	return o.db.GetKeyByIdxKey(idxKey, 0)
}

func (o *BaseObject) KeyToIdxKey(key []byte) ([]byte, error) {

	lenKey := len(key)
	if lenKey != pttdb.SizeDBKeyPrefix+types.SizePttID+types.SizeTimestamp+types.SizePttID {
		return nil, ErrInvalidKey
	}

	idxKey := make([]byte, pttdb.SizeDBKeyPrefix+types.SizePttID+types.SizePttID)

	// prefix
	idxOffset := 0
	nextIdxOffset := pttdb.SizeDBKeyPrefix
	copy(idxKey[:nextIdxOffset], DBOpKeyIdxPrefix)

	// entity-id
	idxOffset = nextIdxOffset
	nextIdxOffset += types.SizePttID

	keyOffset := pttdb.SizeDBKeyPrefix
	nextKeyOffset := keyOffset + types.SizePttID
	copy(idxKey[idxOffset:nextIdxOffset], key[keyOffset:nextKeyOffset])

	// id
	idxOffset = nextIdxOffset
	nextIdxOffset += types.SizePttID

	keyOffset = lenKey - types.SizePttID
	nextKeyOffset = lenKey
	copy(idxKey[idxOffset:nextIdxOffset], key[keyOffset:nextKeyOffset])

	return idxKey, nil
}

func (k *BaseObject) DeleteKey(key []byte) error {
	idxKey, err := k.KeyToIdxKey(key)
	if err != nil {
		return err
	}

	err = k.db.DeleteAll(idxKey)

	if err != nil {
		return err
	}

	return nil
}

func (o *BaseObject) GetBaseObject() *BaseObject {
	return o
}

func (o *BaseObject) GetBlockInfo() BlockInfo {
	return nil
}

func (o *BaseObject) SetBlockInfo(blockInfo BlockInfo) error {
	return nil
}

func (o *BaseObject) RemoveMeta() {
	return
}
