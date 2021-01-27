package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

func ObjectFromFile(tableName string, filePath string) ([]map[string]interface{}, error) {
	b, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		return nil, fmt.Errorf("couldn't read file %s: %w", filePath, readErr)
	}
	log.Printf("[%s] Read %d bytes from disk (%s).\n", tableName, len(b), filePath)

	jsonMap := make([]map[string](interface{}), 0)
	marshalErr := json.Unmarshal([]byte(b), &jsonMap)
	return jsonMap, marshalErr
}
