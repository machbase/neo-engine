
```sh
curl -o - -X POST http://127.0.0.1:4088/db/query \
    --data-urlencode "q=select * from tagdata" \
    --data-urlencode "limit=1" \
    --data-urlencode "cursor=2" \
    |jq
```

```json
{
  "success": true,
  "reason": "1 records selected",
  "elapse": "814.337µs",
  "cursor": 3,
  "data": [
    [
      "name-09",
      "2022-12-06T09:27:10.474136952+09:00",
      0.9009000062942505,
      "",
      -9223372036854776000,
      "",
      "1ed74fcb-91ed-6bf1-9121-bef8031ae3fb",
      "",
      -9223372036854776000,
      ""
    ]
  ]
}
```