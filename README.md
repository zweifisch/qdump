# qdump

dump amqp queue


declear a queue and bind to exchange

```sh
qdump --exchange exhange --topic topic localhost
```

specify queue name and queue ttl

```sh
qdump --queue test-queue \
    --exchange exhange \
    --topic topic \
    --ttl 3600
```

dump and filtering with [jq](http://stedolan.github.io/jq/)

```sh
qdump --queue some-queue | jq 'select(.fieldname=="value")'
```

get a single field

```sh
qdump --queue some-queue | jq .fieldname
```

publish a message

```sh
cat payload.json | qdump --exchange exchange --topic topic localhost
```
