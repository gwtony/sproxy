package sproxy

import (
	"errors"
)

type ConnDict map[string] *SshConn

func initConnDict() *ConnDict {
	dict := &ConnDict{}
	return dict
}

func (dict *ConnDict) lookUpDict(key string) (*SshConn, error) {
	d := *dict

	if d[key] != nil {
		return d[key], nil
	}

	return nil, errors.New("Key not found")
}

func (dict *ConnDict) updateDict(key string, sconn *SshConn) error {
	d := *dict
	d[key] = sconn

	return nil
}

