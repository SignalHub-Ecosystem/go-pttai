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

package pttdb

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
)

type DBable struct {
	UpdateTS types.Timestamp `json:"UT"`
}

func (d *DBable) Unmarshal(data []byte) error {
	return json.Unmarshal(data, d)
}

type DBStatus struct {
	Status types.Status `json:"S"`
}

type DBWithStatus struct {
	BaseObj  *DBStatus       `json:"b"`
	UpdateTS types.Timestamp `json:"UT"`
}

func (d *DBWithStatus) Unmarshal(data []byte) error {
	return json.Unmarshal(data, d)
}