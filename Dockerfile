FROM golang:1.10.3-stretch

COPY . /go/src/github.com/miatachallenge/sync
RUN go install -v github.com/miatachallenge/sync

FROM debian:stretch
COPY --from=0 /go/bin/sync /usr/bin/sync
CMD [ "sync", "--rdb_addr=rethinkdb:28015" ]
