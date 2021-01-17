package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Print(err)
	}

	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, b)
	fmt.Printf("%v", buf)
}
