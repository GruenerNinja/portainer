# Learn Go by exploring Portainer

This guide assumes you know Java but are new to Go. It gives you a route through
the repository, explains the Go syntax you will meet, and follows one small
feature from the browser to the database.

Do not try to understand every package first. Portainer is a mature application
with many integrations. Learn one request path, make a small tested change, and
then widen your view.

## What this project contains

Portainer is one application with two major parts:

- `api/` and `pkg/`: the Go backend. It provides the HTTP API, stores Portainer
  data, and communicates with Docker, Kubernetes, registries, Git, and agents.
- `app/`: the browser frontend. Older screens use AngularJS, while newer screens
  and migrated components use React and TypeScript.

Important supporting directories and files:

| Path                        | Purpose                                              |
| --------------------------- | ---------------------------------------------------- |
| `api/cmd/portainer/main.go` | Backend entry point and dependency wiring            |
| `api/http/server.go`        | Builds handlers and starts HTTP/HTTPS servers        |
| `api/http/handler/`         | API routes grouped by feature                        |
| `api/dataservices/`         | Typed access to stored Portainer data                |
| `api/portainer.go`          | Many shared domain types and interfaces              |
| `api/internal/`             | Backend code intended only for this module           |
| `pkg/`                      | Reusable Go packages with narrower responsibilities  |
| `app/react/`                | React components and modern frontend code            |
| `app/portainer/`            | Portainer UI modules, services, and legacy screens   |
| `Makefile`                  | Common build, test, format, and development commands |
| `go.mod`                    | Go module name, Go version, and backend dependencies |
| `package.json`              | Frontend scripts and dependencies                    |

The broad request flow is:

```text
Browser component
  -> frontend query or mutation
  -> Axios sends /api/... request
  -> top-level Go HTTP handler
  -> feature router and security bouncer
  -> endpoint function
  -> data/service layer or external platform
  -> JSON response
  -> frontend cache and component update
```

## Start here

Install and run the project by following [Development Setup](development-setup.md).
The shortest commands are:

```sh
make deps
make dev
```

Useful URLs are:

- Portainer backend: <https://localhost:9443>
- Frontend development server: <http://localhost:8999>

Before changing anything, run a small backend package test:

```sh
go test ./api/http/handler/tags
```

This is faster and easier to diagnose than running the entire suite. After a Go
change, format and rerun the affected package:

```sh
gofmt -w path/to/changed_file.go
go test ./path/to/changed/package
```

Use these broader checks when the focused test passes:

```sh
make test-server
make lint-server
```

Run `make help` to see all supported commands.

## Go translated into Java terms

Go and Java are both statically typed, but Go favors small interfaces,
composition, explicit errors, and simple control flow.

| Go                             | Rough Java comparison                                   |
| ------------------------------ | ------------------------------------------------------- |
| `package tags`                 | `package ...tags;`                                      |
| `import "net/http"`            | `import java.net.http...;`                              |
| `func main()`                  | `public static void main(String[] args)`                |
| `func add(a int, b int) int`   | `int add(int a, int b)`                                 |
| `name := "dev"`                | `var name = "dev";` with local type inference           |
| `var name string`              | `String name;`, initialized to `""` in Go               |
| `type User struct { ... }`     | A simple class/data record                              |
| `type Store interface { ... }` | A Java interface, implemented implicitly                |
| `func (s *Store) Read()`       | An instance method on `Store`                           |
| `&value`                       | A pointer/reference to `value`                          |
| `*User`                        | A pointer that may be `nil`                             |
| `[]Tag`                        | A growable view over an array, similar to `List<Tag>`   |
| `map[int]bool`                 | Similar to `Map<Integer, Boolean>`; often used as a set |
| `defer close()`                | Run at function exit, like a small `finally` block      |
| `go work()`                    | Start `work` concurrently, roughly a lightweight thread |
| `chan T`                       | A typed queue for communication between goroutines      |
| `context.Context`              | Cancellation/deadline propagated through call chains    |
| `if err != nil`                | Explicit exception-like error handling                  |
| `T any`                        | A generic type parameter such as Java's `<T>`           |

### Functions can return more than one value

Go commonly returns a result and an error:

```go
tag, err := store.Tag().Read(id)
if err != nil {
    return err
}
```

The Java equivalent would usually return a value and throw an exception. Go
makes the possible failure visible at the call site. Never ignore an error until
you understand why it is safe to do so.

`:=` declares local variables. `=` assigns variables that already exist:

```go
tag, err := readTag() // declares tag and err
tag, err = readAgain() // assigns both
```

### Structs hold data

```go
type Tag struct {
    ID   TagID
    Name string
}
```

A struct has fields but no constructors built into the language. A function
named `NewSomething` is only a convention:

```go
func NewTag(name string) *Tag {
    return &Tag{Name: name}
}
```

`&Tag{...}` returns a pointer. Go does not require `new` for normal construction.

### Methods use a receiver

```go
func (handler *Handler) tagList(...) { ... }
```

`handler *Handler` is the receiver. It is close to Java's `this`, except the
receiver has an explicit name and type. A pointer receiver allows the method to
work with the original value rather than a copy.

### Interfaces are implemented implicitly

A type implements a Go interface by having the required methods. There is no
`implements` keyword. This makes interfaces easy to place near the code that
uses them and makes test doubles small.

```go
type Reader interface {
    Read(id int) (*Tag, error)
}
```

Any type with exactly that `Read` method satisfies `Reader`.

### Embedding is composition

You will see this in the tag handler:

```go
type Handler struct {
    *mux.Router
    DataStore dataservices.DataStore
}
```

The unnamed `*mux.Router` field is embedded, so `Handler` exposes the router's
methods. This may look like inheritance, but it is composition and method
promotion. The handler still contains a router.

### Zero values replace many constructors

Go initializes values automatically:

- numbers become `0`
- booleans become `false`
- strings become `""`
- pointers, maps, slices, functions, and interfaces become `nil`

A nil map cannot accept writes, so create one first:

```go
endpoints := map[portainer.EndpointID]bool{}
endpoints[1] = true
```

### Visibility uses capitalization

Names beginning with an uppercase letter are exported from their package.
Lowercase names are package-private:

```go
type Handler struct{}        // other packages can use it
type tagCreatePayload struct{} // only package tags can use it
```

There are no `public`, `private`, or `protected` keywords.

### Struct tags are runtime metadata

This is one field followed by backtick metadata:

```go
Name string `json:"Name" validate:"required"`
```

It resembles Java annotations. Libraries use reflection to read it. The `json`
part controls JSON field naming; `validate` describes a validation rule.

## How the backend starts

Open `api/cmd/portainer/main.go` and begin at `main`, near the bottom. Reading
from `main` downward is not useful because most helper functions are declared
above it. Go does not require functions to be declared before they are called.

`main` performs four important actions:

1. Configure logging.
2. Parse command-line flags.
3. Create a cancellation context and call `buildServer`.
4. Call `server.Start`, which blocks while the server runs.

`buildServer` is large because it is the composition root. It creates the
database, authentication, Docker and Kubernetes clients, schedulers, proxies,
and the HTTP server. Think of it as wiring Spring beans manually. Construction
is explicit, so you can search for `NewService` or follow a value directly.

The `context.Context` passed through services carries cancellation. When its
cancel function is called, background work watching `ctx.Done()` can stop:

```go
select {
case <-ctx.Done():
    return
case item := <-work:
    process(item)
}
```

Do not store a context in a struct unless a package has a deliberate lifecycle
reason. Normal request code passes it as the first parameter.

## How HTTP routing works

`api/http/server.go` creates one handler for each feature. It provides each
handler with dependencies such as the data store and authorization service.
It then puts them into `api/http/handler/handler.go`.

That top-level handler implements the standard library interface:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Its `ServeHTTP` switch sends `/api/tags...` to the Tags handler and removes the
`/api` prefix. The feature handler therefore sees `/tags...`.

Middleware wraps handlers. Conceptually:

```text
request
  -> CSRF protection
  -> panic and slow-request logging
  -> offline/admin checks
  -> top-level router
  -> feature access check
  -> endpoint function
```

The wrapper closest to the final handler runs last on the way in and first on
the way out. This is similar to servlet filters.

## Worked feature: Tags

Tags provide a compact vertical slice of the application.

### 1. Route registration

Open `api/http/handler/tags/handler.go`.

`NewHandler` registers three routes:

| Method and path         | Access             | Go method   |
| ----------------------- | ------------------ | ----------- |
| `POST /api/tags`        | Administrator      | `tagCreate` |
| `GET /api/tags`         | Authenticated user | `tagList`   |
| `DELETE /api/tags/{id}` | Administrator      | `tagDelete` |

The `bouncer` applies access checks. `httperror.LoggerHandler` adapts Portainer's
error-returning endpoint functions to Go's standard `http.Handler` behavior and
logs failures consistently.

### 2. Decode and validate input

Open `api/http/handler/tags/tag_create.go`.

The private payload struct describes accepted input:

```go
type tagCreatePayload struct {
    Name string `validate:"required" example:"org/acme"`
}
```

For this request body:

```json
{ "name": "production" }
```

the shared decoder populates `payload.Name` and invokes `Validate`. Invalid
input becomes a `400 Bad Request` instead of reaching the database.

The comments beginning with `// @` are Swagger annotations, not ordinary prose.
`make generate-api` uses them to generate the OpenAPI document and TypeScript
client. Do not directly edit files under
`app/react/portainer/generated-api/portainer/`; regenerate them instead.

### 3. Use a transaction

Creation runs inside:

```go
handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
    tag, err = createTag(tx, payload)
    return err
})
```

The function is a closure, similar to passing a Java lambda. It can assign the
outer `tag` and `err` variables. If it returns an error, the database transaction
rolls back. If it returns nil, the transaction commits.

Use `ViewTx` for a group of reads and `UpdateTx` when any data may change. Keep
all related writes in the same transaction so another goroutine cannot observe
half of an update.

### 4. Apply a business rule

`createTag` reads existing tags and rejects a duplicate name with HTTP 409. It
then builds a `portainer.Tag` and stores it.

The maps on a tag are sets. A key exists when an environment or group has that
tag. Go has no built-in set type, so `map[ID]bool` is a common representation.

### 5. Persist the model

The shared model is in `api/portainer.go`. The persistence implementation is in
`api/dataservices/tag/tag.go`.

`Service` embeds:

```go
dataservices.BaseDataService[portainer.Tag, portainer.TagID]
```

This generic base supplies common CRUD methods. It is roughly comparable to a
Java `Repository<Tag, TagID>`, but through composition. The tag service defines
custom creation because it must assign the database-generated ID.

There are two service forms:

- `Service` opens its own transaction for a standalone operation.
- `ServiceTx` uses a transaction already owned by the caller.

Inside `UpdateTx`, call `tx.Tag()`, not `handler.DataStore.Tag()`. Opening a
separate transaction would break the all-or-nothing behavior and can deadlock in
some database engines.

### 6. Return JSON or an error

`response.TxResponse(w, tag, err)` turns success into JSON and translates a
typed handler error to an HTTP response. Other endpoints use helpers such as
`response.JSON` and `response.TxEmptyResponse`.

Keep transport concerns at the handler boundary:

- handler: HTTP parsing, status codes, and response JSON
- service/helper: business rules
- data service: persistence

This separation makes the business behavior easier to test without a live
server.

### 7. Fetch it in the frontend

Open `app/portainer/tags/tags.service.ts`. `getTags` and `createTag` use the
shared Axios client. Its base configuration makes `'/tags'` reach
`'/api/tags'` on the backend.

Open `app/portainer/tags/queries.ts`. React Query adds caching and mutation
state. `useCreateTagMutation` invalidates `['tags']` after success, causing
components using `useTags` to receive fresh data.

Finally, `app/react/portainer/environments/TagsView/TagsDatatable.tsx` renders
the supplied tag array. Portainer is migrating gradually from AngularJS, so a
React component may be registered inside an AngularJS route. Do not assume that
every screen has a pure React entry point.

## Reading unfamiliar Go code

Use this repeatable method:

1. Find the route or caller with `rg`.
2. Read the target function's signature and return values.
3. Identify interfaces used by the function.
4. Find the concrete implementation only when its behavior matters.
5. Read nearby tests; they often describe the contract more clearly than the
   implementation.
6. Run that package's tests before and after a small change.

Examples:

```sh
rg -n 'Handle\("/tags' api
rg -n 'func \(handler \*Handler\) tagCreate' api
rg -n 'type TagService interface' api
rg -n 'func .* Create\(tag \*portainer.Tag\)' api
```

In Go, package names matter more than directory-wide class hierarchies. When
you see `response.JSON`, `response` is the imported package and `JSON` is its
exported function.

Your editor's “Go to definition,” “Find references,” and type display are very
useful. Install `gopls`, the official Go language server, if the editor does not
already manage it.

## Testing patterns in this repository

Go tests live beside the code and end in `_test.go`. A function beginning with
`Test` is discovered automatically:

```go
func TestSomething(t *testing.T) { ... }
```

Common tools here include:

- `httptest.NewRequest`: creates an HTTP request without a real network socket.
- `httptest.NewRecorder`: records status, headers, and response body.
- `require`: stops the current test immediately when an assertion fails.
- `assert`: records a failure and lets the test continue.
- `t.Run`: creates named subtests.
- `t.Parallel`: permits independent tests to run concurrently.

Read `api/http/handler/tags/tag_delete_test.go` after the implementation. It
shows how to build a test store, register a handler, send a request, and verify
both the HTTP result and stored data.

Run one test by name:

```sh
go test ./api/http/handler/tags -run TestHandler_tagDelete -v
```

Use the race detector when changing concurrent code:

```sh
go test -race ./path/to/package
```

## Making your first backend change

A safe first exercise is adding a focused validation rule to a small feature.
Use this workflow:

1. Find or add a failing table-driven test.
2. Make the smallest code change that passes it.
3. Run `gofmt` on changed Go files.
4. Run the package test.
5. Run the broader server checks before opening a pull request.

A typical table-driven test looks like this:

```go
func TestValidateName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {name: "valid", input: "production", wantErr: false},
        {name: "empty", input: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateName(tt.input)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

This replaces the parameterized-test patterns you may know from JUnit.

## Adding an API endpoint

When a feature needs a new endpoint, follow the surrounding package rather than
inventing a new architecture:

1. Add a private handler method in the relevant `api/http/handler/<feature>`
   package.
2. Register its route and access policy in that package's `handler.go`.
3. Decode route values and JSON with helpers from `pkg/libhttp/request`.
4. Put related database work in one `ViewTx` or `UpdateTx` callback.
5. Return responses with helpers from `pkg/libhttp/response`.
6. Add Swagger annotations and handler tests.
7. Run `make generate-api` after changing the API contract.
8. Import the generated frontend types/client where appropriate.

Choose the access policy deliberately. Public, authenticated, and administrator
routes have different security consequences; copying the wrong neighboring
route is not safe.

## Common beginner mistakes

### Treating errors like exceptions

Do not let code continue after a meaningful error. Handle it, add context, or
return it. Avoid logging and returning the same error at every layer because
that produces duplicate log messages.

### Confusing nil and empty collections

Both may have length zero, but JSON can encode a nil slice as `null` and an empty
slice as `[]`. APIs often require one specific shape. Follow existing response
types and tests.

### Copying a mutex or stateful struct

Methods on services generally use pointer receivers for a reason. Passing a
stateful value by copy can duplicate locks or state. Let `go vet` and the linter
help catch this.

### Starting a goroutine without an owner

Every goroutine should have a clear stop condition and error strategy. For
long-lived work, look for a context or shutdown channel. Tests otherwise leak
background work and servers may not shut down cleanly.

### Editing generated frontend files

Files under `app/react/portainer/generated-api/portainer/` are overwritten.
Change Swagger annotations or generator configuration, then run
`make generate-api`.

### Commenting syntax instead of intent

Useful comments explain ownership, a business rule, a security choice, or why
an unusual operation exists. A comment such as “increment i” above `i++` adds
noise. The beginner comments added along this learning path are concentrated at
architectural boundaries for that reason.

## Suggested learning path

Work through these in order:

1. Run `go test ./api/http/handler/tags`.
2. Trace `GET /api/tags` from `handler.go` to `tag_list.go` and the data service.
3. Trace `POST /api/tags`, including validation and the transaction.
4. Read the delete test, then the more complex delete implementation.
5. Add one test-only case locally and predict the result before running it.
6. Explore `api/http/handler/system/status.go`, a small public read endpoint.
7. Explore a feature you care about and draw the same route-to-database map.
8. Make a small behavior change with a focused test.

You do not need to memorize Go before contributing. Learn each language feature
when the code gives it a purpose, and keep changes small enough that tests can
confirm your understanding.
