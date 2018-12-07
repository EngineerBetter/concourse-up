package bosh

import (
	"fmt"
	"strings"
)

func vars(vars map[string]interface{}) []string {
	var x []string
	for k, v := range vars {
		switch v.(type) {
		case string:
			if k == "tags" {
				x = append(x, "--var", fmt.Sprintf("%s=%s", k, v))
				continue
			}
			x = append(x, "--var", fmt.Sprintf("%s=%q", k, v))
		case int:
			x = append(x, "--var", fmt.Sprintf("%s=%d", k, v))
		default:
			panic("unsupported type")
		}
	}
	return x
}

type temporaryStore map[string][]byte

func (s temporaryStore) Set(key string, value []byte) error {
	s[key] = value
	return nil
}

func (s temporaryStore) Get(key string) ([]byte, error) {
	return s[key], nil
}

func splitTags(ts []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, t := range ts {
		ss := strings.SplitN(t, "=", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("could not split tag %q", t)
		}
		m[ss[0]] = ss[1]
	}
	return m, nil
}
