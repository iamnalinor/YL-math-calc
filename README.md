# YL-math-calc

## Final Yandex.Lyceum project

Task: Write a program that can calculate the value of a mathematical expression. But there's a catch: every operation takes a lot of time to be completed, say, 1 minute. The program should use Go parallelism to calculate the expression faster.

## Usage

### Setup

1. Install Go from [the official website](https://golang.org/dl/).
2. Clone the repository or download the source code.
3. Install dependencies by running `go mod tidy`.
4. Run `go run cmd/main.go` in the project directory.

The recommended Go version is 1.22. It should work on 1.21 too, however, it is not guaranteed.

Web server will start at [localhost:8081](http://localhost:8081).

### Database

SQLite is used as the database. The database file is db.sqlite3. It is created automatically when the program is run.

### Authorization

Create account:

POST `http://localhost:8081/register`

Body:

```json
{
  "login": "your_login",
  "password": "your_password"
}
```

Curl example:

```bash
curl -X POST http://localhost:8081/api/v1/register -d "{\"login\": \"your_login\", \"password\": \"your_password\"}"
```

Login:

POST `http://localhost:8081/login`

Body:

```json
{
  "login": "your_login",
  "password": "your_password"
}
```

Result will be json with token:

```json
{
  "token": "<token>"
}
```

Curl example:

```bash
curl -X POST http://localhost:8081/api/v1/login -d "{\"login\": \"your_login\", \"password\": \"your_password\"}"
```

Token is valid for 24 hours. **Add the token to the Authorization Bearer header in the following requests in
format `Authorization: Bearer <token>`**.

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
curl -X POST http://localhost:8081/api/v1/createExpression  -H "Authorization: Bearer <token>" -d "{\"expression\": \"2+2*2\"}"
```

Replace `<token>` with the token obtained from the `/login` endpoint.

Examples of expressions:

- `2+2*2`
- `(5 + 5) / (8 * 3)`
- `(0 / 0) + 1`

If one of the operations results with error, the entire expression will be marked as errored.

### Getting result

GET `http://localhost:8081/api/v1/expression/42`

Replace 42 with the ID of the expression. You can obtain the ID from the /createExpression response.

Result:

```json
{
  "id": 42,
  "type": "expression",
  "expression": "2+2*2",
  "status": "done",
  "result": 6,
  "created_time": "2021-10-10T12:00:00Z",
  "finished_time": "2021-10-10T12:01:00Z"
}
```

Curl example:
```bash
curl http://localhost:8081/expression/42 -H "Authorization: Bearer <token>"
```

Replace `<token>` with the token obtained from the `/login` endpoint.

## Docs
Documentation is available at [GitHub Wiki](https://github.com/iamnalinor/YL-math-calc/wiki/Docs).

## License
This code is licensed under the GNU GPL v3.0. You can find the license in the LICENSE file.

## Contact

If you encounter any issues, feel free to contact me at Telegram: [@nalinor](https://t.me/nalinor)
