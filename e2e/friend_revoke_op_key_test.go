// Copyright 2019 The go-pttai Authors
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

package e2e

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/friend"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestFriendRevokeOpKey(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_1.Status)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	//nodeID1_1 := me1_1.NodeID
	//pubKey1_1, _ := nodeID1_1.Pubkey()
	// nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_3 := &me.MyInfo{}
	testCore(t0, bodyString, me0_3, t, isDebug)
	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))

	// 5. show-url
	bodyString = `{"id": "testID", "method": "me_showURL", "params": []}`

	dataShowURL1_5 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowURL1_5, t, isDebug)
	url1_5 := dataShowURL1_5.URL

	// 7. join-friend
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinFriend", "params": ["%v"]}`, url1_5)

	dataJoinFriend0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinFriend0_7, t, isDebug)

	assert.Equal(me1_3.ID, dataJoinFriend0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinFriend0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(TimeSleepDefault)

	// 8. get-friend-list
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getFriendList", "params": ["", 0]}`)

	dataGetFriendList0_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetFriendList0_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList0_8.Result))
	friend0_8 := dataGetFriendList0_8.Result[0]
	assert.Equal(types.StatusAlive, friend0_8.Status)
	assert.Equal(me1_1.ID, friend0_8.FriendID)

	dataGetFriendList1_8 := &struct {
		Result []*friend.BackendGetFriend `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetFriendList1_8, t, isDebug)
	assert.Equal(1, len(dataGetFriendList1_8.Result))
	friend1_8 := dataGetFriendList1_8.Result[0]
	assert.Equal(types.StatusAlive, friend1_8.Status)
	assert.Equal(me0_1.ID, friend1_8.FriendID)
	assert.Equal(friend0_8.ID, friend1_8.ID)

	// 9. get-raw-friend
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getRawFriend", "params": ["%v"]}`, string(marshaled))

	friend0_9 := &friend.Friend{}
	testCore(t0, bodyString, friend0_9, t, isDebug)
	assert.Equal(friend0_8.ID, friend0_9.ID)
	assert.Equal(me1_1.ID, friend0_9.FriendID)
	assert.Equal(me0_1.ID, friend0_9.OwnerIDs[0])

	friend1_9 := &friend.Friend{}
	testCore(t1, bodyString, friend1_9, t, isDebug)
	assert.Equal(friend1_8.ID, friend1_9.ID)
	assert.Equal(friend0_9.Friend0ID, friend1_9.Friend0ID)
	assert.Equal(friend0_9.Friend1ID, friend1_9.Friend1ID)
	assert.Equal(me0_1.ID, friend1_9.FriendID)
	assert.Equal(me1_1.ID, friend1_9.OwnerIDs[0])

	// 9.1. get-raw-board
	marshaled, _ = me0_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaled))

	board0_9_1 := &content.Board{}
	testCore(t0, bodyString, board0_9_1, t, isDebug)
	assert.Equal(me0_1.ID, board0_9_1.CreatorID)
	assert.Equal(me0_1.ID, board0_9_1.OwnerIDs[0])

	board1_9_1 := &content.Board{}
	testCore(t1, bodyString, board1_9_1, t, isDebug)
	assert.Equal(me0_1.ID, board1_9_1.CreatorID)
	assert.Equal(me1_1.ID, board1_9_1.OwnerIDs[0])

	// 9.2. get-raw-board
	marshaled, _ = me1_3.BoardID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaled))

	board0_9_2 := &content.Board{}
	testCore(t0, bodyString, board0_9_2, t, isDebug)
	assert.Equal(me1_1.ID, board0_9_2.CreatorID)
	assert.Equal(me0_1.ID, board0_9_2.OwnerIDs[0])

	board1_9_2 := &content.Board{}
	testCore(t1, bodyString, board1_9_2, t, isDebug)
	assert.Equal(me1_1.ID, board1_9_2.CreatorID)
	assert.Equal(me1_1.ID, board1_9_2.OwnerIDs[0])

	// 9.3. get-raw-account
	marshaled, _ = me0_3.ProfileID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_9_3 := &account.Profile{}
	testCore(t0, bodyString, profile0_9_3, t, isDebug)
	assert.Equal(me0_1.ID, profile0_9_3.CreatorID)
	assert.Equal(me0_1.ID, profile0_9_3.OwnerIDs[0])

	profile1_9_3 := &account.Profile{}
	testCore(t1, bodyString, profile1_9_3, t, isDebug)
	assert.Equal(me0_1.ID, profile1_9_3.CreatorID)
	assert.Equal(me1_1.ID, profile1_9_3.OwnerIDs[0])

	// 9.4. get-raw-account
	marshaled, _ = me1_3.ProfileID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "account_getRawProfile", "params": ["%v"]}`, string(marshaled))

	profile0_9_4 := &account.Profile{}
	testCore(t0, bodyString, profile0_9_4, t, isDebug)
	assert.Equal(me1_1.ID, profile0_9_4.CreatorID)
	assert.Equal(me0_1.ID, profile0_9_4.OwnerIDs[0])

	profile1_9_4 := &account.Profile{}
	testCore(t1, bodyString, profile1_9_4, t, isDebug)
	assert.Equal(me1_1.ID, profile1_9_4.CreatorID)
	assert.Equal(me1_1.ID, profile1_9_4.OwnerIDs[0])

	// 10. master-oplog
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterOplogList0_10 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogList0_10, t, isDebug)
	assert.Equal(2, len(dataMasterOplogList0_10.Result))
	masterOplog0_10_0 := dataMasterOplogList0_10.Result[0]
	masterOplog0_10_1 := dataMasterOplogList0_10.Result[1]
	assert.Equal(types.StatusAlive, masterOplog0_10_0.ToStatus())
	assert.Equal(types.StatusAlive, masterOplog0_10_1.ToStatus())
	assert.Equal(masterOplog0_10_0.ObjID, me1_1.ID)
	assert.Equal(masterOplog0_10_1.ObjID, me0_1.ID)

	dataMasterOplogList1_10 := &struct {
		Result []*pkgservice.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogList1_10, t, isDebug)
	assert.Equal(2, len(dataMasterOplogList1_10.Result))
	assert.Equal(dataMasterOplogList0_10, dataMasterOplogList1_10)
	masterOplog1_10_0 := dataMasterOplogList1_10.Result[0]
	masterOplog1_10_1 := dataMasterOplogList1_10.Result[1]
	assert.Equal(types.StatusAlive, masterOplog1_10_0.ToStatus())
	assert.Equal(types.StatusAlive, masterOplog1_10_1.ToStatus())
	assert.Equal(masterOplog1_10_0.ID, masterOplog1_10_1.MasterLogID)
	assert.Equal(1, len(masterOplog1_10_0.MasterSigns))
	masterSign1_10_0_0 := masterOplog1_10_0.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_10_0_0.ID)
	masterSign1_10_1_0 := masterOplog1_10_1.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_10_1_0.ID)

	// 11. masters
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterListFromCache", "params": ["%v"]}`, string(marshaled))

	dataMasterList0_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11, t, isDebug)
	assert.Equal(2, len(dataMasterList0_11.Result))

	dataMasterList1_11 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterList1_11, t, isDebug)
	assert.Equal(2, len(dataMasterList1_11.Result))

	// 11.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMasterList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMasterList0_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterList0_11_1, t, isDebug)
	assert.Equal(2, len(dataMasterList0_11_1.Result))

	dataMasterList1_11_1 := &struct {
		Result []*pkgservice.Master `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterList1_11_1, t, isDebug)
	assert.Equal(2, len(dataMasterList1_11_1.Result))
	assert.Equal(dataMasterList0_11_1, dataMasterList1_11_1)

	// 12. member-oplog
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMemberOplogList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberOplogList0_12 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberOplogList0_12, t, isDebug)
	assert.Equal(2, len(dataMemberOplogList0_12.Result))
	memberOplog0_12_0 := dataMemberOplogList0_12.Result[0]
	memberOplog0_12_1 := dataMemberOplogList0_12.Result[1]
	assert.Equal(types.StatusAlive, memberOplog0_12_0.ToStatus())
	assert.Equal(types.StatusAlive, memberOplog0_12_1.ToStatus())
	assert.Equal(memberOplog0_12_0.ObjID, me1_1.ID)
	assert.Equal(memberOplog0_12_1.ObjID, me0_1.ID)

	dataMemberOplogList1_12 := &struct {
		Result []*pkgservice.MemberOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberOplogList1_12, t, isDebug)
	assert.Equal(2, len(dataMemberOplogList1_12.Result))
	assert.Equal(dataMemberOplogList0_12, dataMemberOplogList1_12)
	memberOplog1_12_0 := dataMemberOplogList1_12.Result[0]
	memberOplog1_12_1 := dataMemberOplogList1_12.Result[1]
	assert.Equal(types.StatusAlive, memberOplog1_12_0.ToStatus())
	assert.Equal(types.StatusAlive, memberOplog1_12_1.ToStatus())
	assert.Equal(masterOplog0_10_0.ID, memberOplog1_12_0.MasterLogID)
	assert.Equal(masterOplog0_10_0.ID, memberOplog1_12_1.MasterLogID)
	assert.Equal(1, len(memberOplog1_12_0.MasterSigns))
	masterSign1_12_0_0 := memberOplog1_12_0.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_12_0_0.ID)
	masterSign1_12_1_0 := memberOplog1_12_1.MasterSigns[0]
	assert.Equal(me1_1.ID, masterSign1_12_1_0.ID)

	// 12.1
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getMemberList", "params": ["%v", "", 0, 2]}`, string(marshaled))

	dataMemberList0_12_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMemberList0_12_1, t, isDebug)
	assert.Equal(2, len(dataMemberList0_12_1.Result))

	dataMemberList1_12_1 := &struct {
		Result []*pkgservice.Member `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMemberList1_12_1, t, isDebug)
	assert.Equal(2, len(dataMemberList1_12_1.Result))
	assert.Equal(dataMemberList0_12_1, dataMemberList1_12_1)

	// 13.
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getOpKeyInfos", "params": ["%v"]}`, string(marshaled))

	dataOpKeyList0_13 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyList0_13, t, isDebug)
	assert.Equal(1, len(dataOpKeyList0_13.Result))
	opKey0_13 := dataOpKeyList0_13.Result[0]

	dataOpKeyList1_13 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t1, bodyString, dataOpKeyList1_13, t, isDebug)
	assert.Equal(1, len(dataOpKeyList1_13.Result))
	opKey1_13 := dataOpKeyList0_13.Result[0]

	assert.Equal(opKey0_13, opKey1_13)

	// 14. t1 shutdown

	bodyString = `{"id": "testID", "method": "ptt_shutdown", "params": []}`

	resultString := `{"jsonrpc":"2.0","id":"testID","result":true}`
	testBodyEqualCore(t1, bodyString, resultString, t)

	time.Sleep(5 * time.Second)

	// 14.1. test-error
	err1_14_1 := testError("http://127.0.0.1:9451")
	assert.NotEqual(nil, err1_14_1)

	// 15. offsetSecond
	offsetSecond := int64(86400 * 8)
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "ptt_setOffsetSecond", "params": [%v]}`, offsetSecond)

	bool0_15 := false
	testCore(t0, bodyString, &bool0_15, t, isDebug)

	// 16. sync.
	t.Logf("16. force sync")
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_forceSync", "params": ["%v"]}`, string(marshaled))

	bool0_16 := false
	testCore(t0, bodyString, &bool0_16, t, isDebug)

	time.Sleep(5 * time.Second)

	// 17.
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getOpKeyInfos", "params": ["%v"]}`, string(marshaled))

	dataOpKeyList0_17 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyList0_17, t, isDebug)
	assert.Equal(0, len(dataOpKeyList0_17.Result))

	// 18. force op-key.
	t.Logf("18. force sync")
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_forceOpKey", "params": ["%v"]}`, string(marshaled))

	bool0_18 := false
	testCore(t0, bodyString, &bool0_18, t, isDebug)

	time.Sleep(5 * time.Second)

	// 19.
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getOpKeyInfos", "params": ["%v"]}`, string(marshaled))

	dataOpKeyList0_19 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyList0_19, t, isDebug)
	assert.Equal(1, len(dataOpKeyList0_19.Result))
	opKey0_19 := dataOpKeyList0_19.Result[0]

	// 20. restart node 1
	startNode(t, 1, offsetSecond, false)

	time.Sleep(TimeSleepRestart)

	// 21.
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getOpKeyInfos", "params": ["%v"]}`, string(marshaled))

	dataOpKeyList0_21 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyList0_21, t, isDebug)
	assert.Equal(2, len(dataOpKeyList0_21.Result))
	opKey0_21_0 := dataOpKeyList0_21.Result[0]
	opKey0_21_1 := dataOpKeyList0_21.Result[1]

	var opKey0_21_0_0 *pkgservice.KeyInfo
	var opKey0_21_0_1 *pkgservice.KeyInfo
	if reflect.DeepEqual(opKey0_21_0.ID, opKey0_19.ID) {
		opKey0_21_0_0 = opKey0_21_0
		opKey0_21_0_1 = opKey0_21_1
	} else {
		opKey0_21_0_0 = opKey0_21_1
		opKey0_21_0_1 = opKey0_21_0
	}

	assert.Equal(opKey0_19, opKey0_21_0_0)

	dataOpKeyList1_21 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t1, bodyString, dataOpKeyList1_21, t, isDebug)
	assert.Equal(2, len(dataOpKeyList1_21.Result))
	opKey1_21_0 := dataOpKeyList1_21.Result[0]
	opKey1_21_1 := dataOpKeyList1_21.Result[1]

	var opKey1_21_0_0 *pkgservice.KeyInfo
	var opKey1_21_0_1 *pkgservice.KeyInfo
	if reflect.DeepEqual(opKey1_21_0.ID, opKey0_19.ID) {
		opKey1_21_0_0 = opKey1_21_0
		opKey1_21_0_1 = opKey1_21_1
	} else {
		opKey1_21_0_0 = opKey1_21_1
		opKey1_21_0_1 = opKey1_21_0
	}

	assert.Equal(opKey0_19, opKey1_21_0_0)
	assert.Equal(opKey0_21_0_1, opKey1_21_0_1)

	// 22. force op-key.
	t.Logf("22. force sync")
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_forceOpKey", "params": ["%v"]}`, string(marshaled))

	bool0_22 := false
	testCore(t0, bodyString, &bool0_22, t, isDebug)

	time.Sleep(10 * time.Second)

	// 23.
	marshaled, _ = friend0_8.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "friend_getOpKeyInfos", "params": ["%v"]}`, string(marshaled))

	dataOpKeyList0_23 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t0, bodyString, dataOpKeyList0_23, t, isDebug)
	assert.Equal(3, len(dataOpKeyList0_23.Result))

	dataOpKeyList1_23 := &struct {
		Result []*pkgservice.KeyInfo `json:"result"`
	}{}
	testListCore(t1, bodyString, dataOpKeyList1_23, t, isDebug)
	assert.Equal(3, len(dataOpKeyList1_23.Result))
}
