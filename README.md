# MiataChallenge Sync

An endpoint that accepts events from the
[bridge](https://github.com/miatachallenge/bridge) and saves them into
a RethinkDB database. Accepts `-rdb_addr` - the RethinkDB server's IP as a flag.
It creates tables and indexes in the `mctracker` database.

You have to manually fill the `keys` table with your own keys using the following
structure:

```javascript
{
	"id": "the key",
	"name": "its description shown in the Processing server's UI"
}
```

Run it using Docker (there's a `Dockerfile` in the repo) or just by compiling
it yourself using Go >=1.8.
