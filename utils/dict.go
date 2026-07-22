package utils

import (
	"fmt"

	"github.com/mo7amed4522/tayara_api/global"
	"github.com/mo7amed4522/tayara_api/model/bxapi"
)

// DictUtils Dictionary utility class
type DictUtils struct{}

var Dict = new(DictUtils)

// GetDictValue Get the first value based on the dictionary type
// dictType is the dictionary type, using constants defined in constants.
func (d *DictUtils) GetDictValue(dictType string) (string, error) {
	if global.GVA_DB == nil {
		return "", fmt.Errorf("database not initialized")
	}

	var dict bxapi.BxDict
	err := global.GVA_DB.Where("type = ? AND is_deleted = ?", dictType, false).
		Order("sort ASC, id ASC").
		First(&dict).Error

	if err != nil {
		return "", fmt.Errorf("failed to get dict value: %v", err)
	}

	return dict.Value, nil
}

// GetDictValues Batch get dictionary values
// dictTypes is the list of dictionary types to resolve.
func (d *DictUtils) GetDictValues(dictTypes []string) (map[string]string, error) {
	if global.GVA_DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var dicts []bxapi.BxDict
	err := global.GVA_DB.Where("type IN ? AND is_deleted = ?", dictTypes, false).
		Order("sort ASC, id ASC").
		Find(&dicts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get dict values: %v", err)
	}

	result := make(map[string]string)
	for _, dict := range dicts {
		result[dict.Type] = dict.Value
	}

	// Ensure that all requested types have a return value, and missing ones are set to an empty string
	for _, dictType := range dictTypes {
		if _, exists := result[dictType]; !exists {
			result[dictType] = ""
		}
	}

	return result, nil
}
