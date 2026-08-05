package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JSONMap map[string]interface{}

// GormDataType lets AutoMigrate map JSONMap to a JSON column on both MySQL and embedded
// drivers (sqlite) instead of failing with "unsupported data type".
func (JSONMap) GormDataType() string {
	return "json"
}

// GormDBDataType refines the column type per dialect (kept for MySQL JSON support).
func (JSONMap) GormDBDataType(_ *gorm.DB, _ *schema.Field) string {
	return "json"
}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}
	var err error
	switch value := value.(type) {
	case []byte:
		err = json.Unmarshal(value, m)
	case string:
		err = json.Unmarshal([]byte(value), m)
	default:
		err = errors.New("basetypes.JSONMap.Scan: invalid value type")
	}
	if err != nil {
		return err
	}
	return nil
}

type TreeNode[T any] interface {
	GetChildren() []T
	SetChildren(children T)
	GetID() int
	GetParentID() int
}
