package rethinkdb

import (
	"github.com/pkg/errors"
	r "gopkg.in/gorethink/gorethink.v4"
)

// Manifest is the main struct of a RethinkDB schema
type Manifest struct {
	Tables []Table
}

func (m Manifest) tableNames() []string {
	result := []string{}
	for _, table := range m.Tables {
		result = append(result, table.Name)
	}
	return result
}

// Table is an element representing a single table in the schema.
type Table struct {
	Name    string
	Indexes map[string]Index
}

/*
	func (t Table) indexNames() []string {
		result := []string{}
		for name := range t.Indexes {
			result = append(result, name)
		}
		return result
	}
*/

// Index is an element representing a single index in the schema.
type Index struct {
	Value interface{}
	Opts  *r.IndexCreateOpts
}

// Prepare prepares a RethinkDB database according to the passed schema manifest.
func Prepare(db *r.Session, man Manifest) error {
	if err := r.Expr(man.tableNames()).ForEach(func(item r.Term) r.Term {
		return r.TableList().Contains(item).Branch(
			map[string]interface{}{},
			r.TableCreate(item),
		)
	}).Exec(db); err != nil {
		return errors.Wrap(err, "uanble to create tables")
	}

	for _, table := range man.Tables {
		for name, index := range table.Indexes {
			var createFunc r.Term

			if index.Value == nil {
				if index.Opts == nil {
					createFunc = r.Table(table.Name).IndexCreate(name)
				} else {
					createFunc = r.Table(table.Name).IndexCreate(name, *index.Opts)
				}
			} else {
				switch value := index.Value.(type) {
				case string:
					if index.Opts == nil {
						createFunc = r.Table(table.Name).IndexCreate(value)
					} else {
						createFunc = r.Table(table.Name).IndexCreate(value, *index.Opts)
					}
				default:
					if index.Opts == nil {
						createFunc = r.Table(table.Name).IndexCreateFunc(name, value)
					} else {
						createFunc = r.Table(table.Name).IndexCreateFunc(name, value, *index.Opts)
					}
				}
			}

			if err := r.Table(table.Name).IndexList().Contains(name).Branch(
				map[string]interface{}{},
				createFunc,
			).Exec(db); err != nil {
				return errors.Wrapf(err, "unable to create index %s of table %s", name, table.Name)
			}
		}
	}

	return nil
}
