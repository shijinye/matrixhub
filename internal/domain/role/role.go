// Copyright The MatrixHub Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package role

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/matrixhub-ai/matrixhub/internal/domain/authz"
)

// RoleType represents role type
type RoleType int

const (
	PlatformRoleAdmin RoleType = iota + 1
	ProjectRoleAdmin
	ProjectRoleEditor
	ProjectRoleViewer
)

// Role represents a role in the system
type Role struct {
	ID          int            `gorm:"primary_key"`
	Name        string         `gorm:"column:name"`
	Permissions PermissionList `gorm:"column:permissions;type:text"`
	Scope       string         `gorm:"column:scope"`
}

func (Role) TableName() string {
	return "roles"
}

// PermissionList is a slice of permissions that can be stored as JSON in database
type PermissionList []authz.Permission

// Value implements driver.Valuer for database storage
func (p PermissionList) Value() (driver.Value, error) {
	if p == nil {
		return "[]", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements sql.Scanner for database retrieval
func (p *PermissionList) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, p)
}
