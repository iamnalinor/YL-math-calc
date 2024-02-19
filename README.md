# YL-math-calc
## 1st semester Yandex.Lyceum project

Task: Write a program that can calculate the value of a mathematical expression. But there's a catch: every operation takes a lot of time to be completed, say, 1 minute. The program should use Go parallelism to calculate the expression faster.

## Usage
### Start
The project itself has no dependencies. The recommended Go version is 1.22. It is not guaranteed to work on older versions.

Clone the repository, **copy `config.json.dist` to `config.json`** and run the main.go file: `go run cmd/main.go`. Web server will start at localhost:8081.

### Creating expression
POST `http://localhost:8081/createExpression`

Body:
```json
{
    "expression": "2+2*2"
}
```

Additionally, you can specify idempotency token in `X-Idempotency-Token`.

Result:

```json
{"id":2}
```

Curl example:
```bash
curl -X POST http://localhost:8081/createExpression -H "Content-Type: application/json" -d "{\"expression\": \"2+2*2\"}"
```

### Getting result
GET `http://localhost:8081/expression/42`

Replace 42 with the ID of the expression. You can obtain the ID from the /createExpression response.

Result:

```json
{
  "id": 42,
  "type": "expression",
  "status": "done",
  "result": 6,
  "created_time": "2021-10-10T12:00:00Z",
  "finished_time": "2021-10-10T12:01:00Z"
}
```

Curl example:
```bash
curl http://localhost:8081/expression/42
```

## Docs
Documentation is available at [GitHub Wiki](https://github.com/iamnalinor/YL-math-calc/wiki/Docs).

## License
This code is licensed under the GNU GPL v3.0. You can find the license in the LICENSE file.