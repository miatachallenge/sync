package main

import (
	"github.com/miatachallenge/sync/rethinkdb"
)

var schema = rethinkdb.Manifest{
	Tables: []rethinkdb.Table{
		{
			Name: "keys",
			Indexes: map[string]rethinkdb.Index{
				"name": {},
			},
		},
		{
			Name: "records",
			Indexes: map[string]rethinkdb.Index{
				"key":       {},
				"timestamp": {},
			},
		},
	},
}
