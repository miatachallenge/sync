package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/miatachallenge/sync/rethinkdb"
	r "gopkg.in/gorethink/gorethink.v4"
)

var (
	bind    = flag.String("bind", ":8090", "bind addr")
	rdbAddr = flag.String("rdb_addr", "127.0.0.1:28015", "rdb address")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Event struct {
	ID        int64  `json:"id" db:"id" gorethink:"internal_id"`
	Key       string `gorethink:"key"`
	Name      string `json:"-" db:"name" gorethink:"name"`
	TagID     string `json:"tag_id" db:"tag_id" gorethink:"tag_id"`
	Timestamp int64  `json:"timestamp" db:"timestamp" gorethink:"timestamp"`
	Antenna   int    `json:"antenna" db:"antenna" gorethink:"antenna"`
}

type Key struct {
	ID   string `gorethink:"id"`
	Name string `gorethink:"name"`
}

func main() {
	flag.Parse()

	db, err := r.Connect(r.ConnectOpts{
		Address:  *rdbAddr,
		Database: "mctracker",
	})
	if err != nil {
		panic(err)
	}

	if err := rethinkdb.Prepare(db, schema); err != nil {
		panic(err)
	}

	http.HandleFunc("/sync", func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		var (
			currentKey   *Key
			currentCount int
		)
		for {
			var input struct {
				Type   string   `json:"type"`
				Key    string   `json:"key"`
				Events []*Event `json:"events"`
			}
			if err := conn.ReadJSON(&input); err != nil {
				conn.WriteJSON(map[string]interface{}{
					"type":  "error",
					"error": err.Error(),
				})
				log.Printf("reading json failed: %s", err)
				return
			}

			if input.Type == "auth" {
				if input.Key == "" {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": "Malformed authentication message",
					})
					continue
				}

				cursor, err := r.Table("keys").Get(input.Key).Run(db)
				if err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}
				key := &Key{}
				if err := cursor.One(key); err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}

				log.Printf("Authenticated user as %s", key.ID)

				currentKey = key

				// Return the current count
				cursor, err = r.Table("records").
					GetAllByIndex("key", currentKey.ID).
					OrderBy(r.Desc("internal_id")).Nth(0).
					Field("internal_id").Default(0).Run(db)
				if err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}
				if err := cursor.One(&currentCount); err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}

				conn.WriteJSON(map[string]interface{}{
					"type":    "current",
					"current": currentCount,
				})

				continue
			}

			if currentKey == nil {
				conn.WriteJSON(map[string]interface{}{
					"type":  "error",
					"error": "Unauthenticated",
				})
				continue
			}

			if input.Type == "sync" {
				for _, event := range input.Events {
					event.Key = currentKey.ID
				}

				result, err := r.Table("records").Insert(input.Events).RunWrite(db)
				if err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}
				log.Printf("Inserted %d records", result.Inserted)

				cursor, err := r.Table("records").
					GetAllByIndex("key", currentKey.ID).
					OrderBy(r.Desc("internal_id")).Nth(0).
					Field("internal_id").Default(0).Run(db)
				if err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}
				if err := cursor.One(&currentCount); err != nil {
					conn.WriteJSON(map[string]interface{}{
						"type":  "error",
						"error": err.Error(),
					})
					continue
				}

				conn.WriteJSON(map[string]interface{}{
					"type":    "current",
					"current": currentCount,
				})
				continue
			}
		}
	})

	http.ListenAndServe(*bind, nil)
}
