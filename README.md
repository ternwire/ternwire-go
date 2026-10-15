# Ternwire Go SDK

Go 1.22+ module with no external dependencies.

```sh
go get github.com/ternwire/ternwire-go
```

```go
import ternwire "github.com/ternwire/ternwire-go"
```

For local development, add a replacement to the consuming application's `go.mod`:

```go
require github.com/ternwire/ternwire-go v0.1.0
replace github.com/ternwire/ternwire-go => /absolute/path/to/ternwire-go
```

## Usage

```go
client, err := ternwire.NewClient(ternwire.Config{
    APIKey: os.Getenv("TERNWIRE_API_KEY"),
    Timeout: 30 * time.Second,
})
if err != nil { return err }
message, err := client.SubmitEmail(ctx, ternwire.Object{
    "from": "notifications@your-verified-domain.com",
    "to": []string{"recipient@example.com"},
    "subject": "Your receipt",
    "text": "Thanks for your order.",
}, "order-123-receipt")
if err != nil { return err }
```

Use the snippet inside a function returning `error`, with `os`, `time`, and a
`context.Context` in scope. `examples/send/main.go` is a complete executable.

The client calls `ternwire.DefaultBaseURL` (`https://ternwire.com`) when `BaseURL`
is empty. Set `BaseURL` only to reach another deployment, such as
`BaseURL: "http://localhost:3001"` for local development. Give the origin without
`/v1`. It must use https unless it points at localhost or a loopback address.

All 17 operations take a context first and return `(Object, error)`, where
`Object` is `map[string]any`. JSON fields stay camelCase. Template deletion returns
`nil, nil`. A zero timeout uses the 30-second default. Reuse the client across calls.
`ListEmailTemplates(ctx, ternwire.ListEmailTemplatesQuery{Cursor: next, Limit: 50})`
returns one page of summaries. Zero fields are left out of the request, so pass
`ListEmailTemplatesQuery{}` for the first page and `page.nextCursor` as `Cursor` after it.

Use `errors.As(err, &apiError)` with `var apiError *ternwire.APIError` to inspect
`Status`, `Body`, `RawBody`, and `Headers`. Ordinary transport errors are preserved.

`ternwire.VerifyWebhook(header, rawBody, secret)` verifies raw bytes with a
five-minute tolerance. Use `VerifyWebhookWithTolerance` to supply a `time.Duration`.

To run the example later, set `TERNWIRE_API_KEY`, `TERNWIRE_FROM`, and `TERNWIRE_TO`,
then run `go run ./examples/send`. It submits an email. Set `TERNWIRE_API_ORIGIN` only
to send it to an origin other than `https://ternwire.com`.

Read the full documentation at [ternwire.com/docs](https://ternwire.com/docs), or contact
support@ternwire.com. Released under the MIT License.
