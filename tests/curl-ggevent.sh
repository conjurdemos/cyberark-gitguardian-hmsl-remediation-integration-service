#!/usr/bin/env bash

export HOST="${1:-http://127.0.0.1:9191}"

if ! command -v ./bin/randchar; then
    echo "Run test script from project base dir"
    exit 1
fi

hashval=$(./bin/randchar -n 64 -a "abcdef0123456789")

# this is a test hmsl hash from gg
#hashval="1da3780a19660a715de68584c2c95a5824968d297606c088d7e05e9f76ab968a"

lt=$(./bin/randchar -n 4 -a "1234567890")
rt=$(./bin/randchar -n 2 -a "1234567890")
acctid="${lt}_${rt}"

jq -n --arg HMSLHASH "$hashval" --arg ACCTID "$acctid" -f ./tests/curl-ggevent-hashes.jq > ./tests/curl-ggevent-hashes.json
jq -n --arg HMSLHASH "$hashval" -f ./tests/ggevent-incident-body.jq > ./tests/ggevent-incident-body.json

curl -D - -X PUT "${HOST}/v1/hashes" \
     -H "Authorization: Bearer dev123" \
     -H "Accept: application/json" \
     -H "Content-Type: application/json" \
     -d "@./tests/curl-ggevent-hashes.json"

curl -D - -X POST "${HOST}/v1/notify/ggevent" \
     -H "Authorization: Bearer dev123" \
     -H "Content-type: application/json" \
     -d "@./tests/ggevent-incident-body.json"

