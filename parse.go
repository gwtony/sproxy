package sproxy

import (
	"encoding/json"
	"io"
	"strings"
)

func parseRequest(data string) (*RequestData, error) {

	var rd RequestData
	dec := json.NewDecoder(strings.NewReader(data))

	err := dec.Decode(&rd)

	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	return &rd, nil
}
