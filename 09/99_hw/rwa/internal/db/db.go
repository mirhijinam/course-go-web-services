package db

import (
	"encoding/json"
)

type DB struct {
	profiles map[string]json.RawMessage
	articles map[string]json.RawMessage
	sessions map[string]json.RawMessage
}

func New() DB {
	return DB{
		profiles: make(map[string]json.RawMessage),
		articles: make(map[string]json.RawMessage),
		sessions: make(map[string]json.RawMessage),
	}
}
