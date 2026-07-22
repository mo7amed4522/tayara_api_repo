package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/mo7amed4522/tayara_api/global"
)

const signkey = "sMDB3mOS00NDk3w8LTfg1"

func Sign(treeMap map[string]interface{}, signKey string) string {
	if signKey == "" {
		signKey = signkey
	}
	var sb strings.Builder

	keys := make([]string, 0, len(treeMap))
	for k := range treeMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := treeMap[key]
		if value != nil {
			URLEncoded := IsURLEncoded(fmt.Sprintf("%v", value))
			if URLEncoded {
				value, _ = url.QueryUnescape(fmt.Sprintf("%v", value))
			}
			sb.WriteString(fmt.Sprintf("%s=%v&", key, value))
		}
	}

	sb.WriteString(fmt.Sprintf("key=%s", signKey))

	// Record the plaintext before the MD5 hash.
	plainText := sb.String()
	global.GVA_LOG.Info("Plaintext before MD5: " + plainText)

	hasher := md5.New()
	hasher.Write([]byte(sb.String()))
	signature := hex.EncodeToString(hasher.Sum(nil))
	// Record the final signature
	global.GVA_LOG.Info("Generated signature: " + strings.ToUpper(signature))

	return strings.ToUpper(signature)
}

func IsURLEncoded(s string) bool {
	// Try to decode
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		// Decoding failed, possibly an unencoded string
		return false
	}

	// Compare the decoded string with the original string
	if decoded != s {
		// The decoded string is different from the original string, indicating that the original string may have been URL encoded
		return true
	}

	// If the decoded string is the same as the original string, it may not be encoded
	return false
}

// Special signature for payment and withdrawal order interfaces
func OrderSign(treeMap map[string]interface{}, signKey string) string {
	if signKey == "" {
		signKey = signkey
	}
	var sb strings.Builder

	keys := make([]string, 0, len(treeMap))
	for k := range treeMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := treeMap[key]
		if value != nil && value != "" {
			URLEncoded := IsURLEncoded(fmt.Sprintf("%v", value))
			if URLEncoded {
				value, _ = url.QueryUnescape(fmt.Sprintf("%v", value))
			}
			sb.WriteString(fmt.Sprintf("%s=%v&", key, value))
		}
	}

	sb.WriteString(fmt.Sprintf("signKey=%s", signKey))

	// Record the plaintext before the MD5 hash
	plainText := sb.String()
	global.GVA_LOG.Info("Plaintext before MD5: " + plainText)

	hasher := md5.New()
	hasher.Write([]byte(sb.String()))
	signature := hex.EncodeToString(hasher.Sum(nil))
	// Record the final signature
	global.GVA_LOG.Info("Generated signature: " + signature)

	return signature
}

func StructToMapNew(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	objValue := reflect.ValueOf(obj).Elem()
	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldInfo := objType.Field(i)
		mapKey := fieldInfo.Name

		// Get the json tag, if it exists, use the tag value
		if tag := fieldInfo.Tag.Get("json"); tag != "" {
			index := strings.Index(tag, ",")
			if index != -1 {
				// Extract the part before the comma
				tag = tag[:index]
			}
			mapKey = tag
		}

		// Convert the first letter of the field name to lowercase
		mapKey = toLowerFirst(mapKey)
		result[mapKey] = field.Interface()
	}

	return result, nil
}

// Convert the first letter of the string to lowercase
func toLowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}
