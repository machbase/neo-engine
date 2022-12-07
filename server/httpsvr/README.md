
```sh
curl -o - -X POST http://127.0.0.1:4088/db/query \
    --data-urlencode "q=select * from tagdata" \
    --data-urlencode "limit=1" \
    --data-urlencode "cursor=2" \
    |jq
```

| param      | default  | desc              |
| ---------- | -------- | ----------------- |
| q          |          | sql text          |
| limit      | 10       | result set limit  |
| cursor     | 0        | result set cursor |
| timeformat | epoch    | format of datetime column           |
|            |          | `epoch`: unix epoch in nano seconds |

```json
{
  "success": true,
  "reason": "1 records selected",
  "elapse": "1.044079ms",
  "cursor": 5,
  "data": {
    "colums": [
      "col01", "col02", "col03", "col04", "col05", "col06", "col07", "col08", "col09", "col10"
    ],
    "types": [
      "string", "datetime", "float64", "string", "int64", "string", "string", "string", "int64", "string"
    ],
    "records": [
      [
        "name-17",
        1670291684975704000,
        1.7017000913619995,
        "",
        null,
        "",
        "1ed7508f-50c2-6b67-aecc-7b54ab426d7a",
        "",
        null,
        ""
      ]
    ]
  }
}
```