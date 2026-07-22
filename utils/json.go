package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func GetJSONKeys(jsonStr string) (keys []string, err error) {
	// Use json.Decoder to record the order of keys during parsing.
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	// Ensure the data is an object
	if t != json.Delim('{') {
		return nil, err
	}
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return nil, err
		}
		keys = append(keys, t.(string))

		// Parse the value
		var value interface{}
		err = dec.Decode(&value)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// PrettyJSON Pretty print
func PrettyJSON(v interface{}) {
	prettyJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshalling failed: %s", err)
	}
	fmt.Println(string(prettyJSON))
}
