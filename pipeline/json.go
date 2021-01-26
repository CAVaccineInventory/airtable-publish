package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
)

func ObjectFromFile(filePath string) ([]map[string]interface{}, error) {
	b, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		return nil, errors.Wrapf(readErr, "couldn't read file %s", filePath)
	}
	log.Printf("Read %d bytes from disk (%s).\n", len(b), filePath)

	jsonMap := make([]map[string](interface{}), 0)
	marshalErr := json.Unmarshal([]byte(b), &jsonMap)
	return jsonMap, marshalErr
}
