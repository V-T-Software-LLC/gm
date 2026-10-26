# gm

gm is a command line tool that scaffolds new projects. You run `gm create`,
answer a few prompts, and get a project that builds and runs, laid out the
way we like to start our projects.

gm currently creates two kinds of projects:

- **Web apps**: a Go server with echo, templ, htmx, and sqlc, with users in
  Postgres and sessions in Redis, plus a Vite Svelte app with Tailwind for the
  parts of the page that need to be dynamic.
- **Low level projects**: C or C++ libraries and applications built with
  CMake, or C projects built with a basic Makefile.

## Installing

gm needs Go 1.23 or newer.

```sh
go install github.com/V-T-Software-LLC/gm@latest
```

This installs `gm` into `$(go env GOPATH)/bin`, which should be on your
`PATH`.

To build from source instead:

```sh
git clone git@github.com:V-T-Software-LLC/gm.git gm
cd gm
make build        # builds bin/gm
```

Put `bin/gm` somewhere on your `PATH`, or run `go install .` from the repo to
install it into your Go bin directory.

## Usage

```sh
gm create
```

gm asks one question at a time. Every project starts with:

1. **Project name**: gm creates a directory with this name in the current
   directory. It fails if the directory already exists.
2. **Project type**: web app or low level.

The remaining questions depend on the project type and are covered below.
While gm works it prints each step, and it only shows a command's output if
that command fails. When it finishes it prints the commands to start working
on the project.

```
✓ Creating demo
✓ Initializing go module
✓ Creating vite svelte app
...

Created demo. To start developing:

  cd demo
  make dev
```

## Web apps

### Prompts

- **Go module name**: passed to `go mod init`, for example
  `github.com/you/demo`.
- **JavaScript package manager**: npm, pnpm, or bun.

### Requirements

These need to be installed when you run `gm create`:

- `go`
- `templ` (`go install github.com/a-h/templ/cmd/templ@latest`)
- `sqlc` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- the package manager you chose (`npm`, `pnpm`, or `bun`)

To run the project you need a Postgres database, a Redis server, and
[goose](https://github.com/pressly/goose) to run migrations
(`go install github.com/pressly/goose/v3/cmd/goose@latest`).

These are optional:

- `air` for live reload during development
  (`go install github.com/air-verse/air@latest`). If gm finds `air` on your
  `PATH` it writes an `.air.toml` and `make dev` uses it. Without it,
  `make dev` runs the server with `go run` and you restart it yourself.

### What gets created

The project follows the
[standard Go project layout](https://github.com/golang-standards/project-layout).

```
demo/
  cmd/demo/main.go            connects to postgres and redis, starts the server
  internal/server/server.go   echo setup, routes, and a templ render helper
  internal/server/auth.go     sign up, sign in, sign out, and requireUser
  internal/server/session.go  sessions stored in redis
  internal/db/migrations/     goose migrations, which sqlc reads as the schema
  internal/db/                queries.sql; sqlc generates here
  web/app/                    vite + svelte + tailwind project
  web/static/                 vite build output, served at /static
  web/views/                  templ templates
  sqlc.yaml                   postgres, pgx/v5
  Makefile
  .env.example                the environment variables the server needs
  .air.toml                   only when air is installed
  .gitignore
```

The server listens on `:8080`:

| Route             | What it does                                       |
| ----------------- | -------------------------------------------------- |
| `GET /`           | home page with a small Svelte counter component    |
| `GET /signup`     | sign up page                                       |
| `POST /signup`    | create a user and a session                        |
| `GET /signin`     | sign in page                                       |
| `POST /signin`    | check the password and create a session            |
| `POST /signout`   | delete the session                                 |
| `GET /dashboard`  | the signed in user's dashboard                     |

#### Users and sessions

The first migration creates a `users` table with first name, last name, email, a bcrypt
password hash, and created and updated timestamps. Emails are lowercased
before they are stored, so they are unique regardless of case.

The sign up and sign in forms post with htmx. When a field is wrong the server
renders the form again with an error, and htmx swaps it in place. On success
the server creates a session and responds with `HX-Redirect: /dashboard`.

A session is a random token in an `HttpOnly`, `SameSite=Lax` cookie, stored in
Redis as `session:<token>` with the user id as the value. Sessions last 30
days. The cookie is `Secure` unless the server runs with `-dev`. To protect a
route, add the `requireUser` middleware. It redirects to `/signin` when there
is no session, and otherwise stores the user in the context:

```go
e.GET("/settings", s.settings, s.requireUser)

user := c.Get("user").(db.User)
```

### Developing

```sh
cp .env.example .env   # set DATABASE_URL and REDIS_URL
make migrate           # create the tables
make dev
```

The Makefile loads `.env` and exports it, so `make dev`, `make run`, and
`make migrate` get the variables. When running the binary yourself, set them
in the environment.

This runs two things in parallel:

- `dev-go`: the Go server with `-dev` (through air when installed). Air runs
  `templ generate` before each build and rebuilds on `.go` and `.templ`
  changes.
- `dev-web`: the Vite dev server on `localhost:5173`.

Open http://localhost:8080. In development the Go server does not serve any
JavaScript or CSS. With `-dev`, the layout loads `@vite/client` and
`src/main.js` straight from the Vite dev server, so Svelte and Tailwind
changes show up immediately with hot reload. This follows Vite's
[backend integration guide](https://vite.dev/guide/backend-integration).

Tailwind scans `web/views` as well as the Svelte sources, so classes used only
in templ files are included.

### Building for production

```sh
make build        # templ generate, vite build, go build -o bin/demo
./bin/demo        # run from the project root
```

`vite build` writes to `web/static` with fixed file names (`main.js` and
`app.css`), so the layout links them directly and there is no manifest to
read. Run the binary from the project root since it serves `web/static` by
relative path.

### Makefile targets

| Target          | What it does                                    |
| --------------- | ----------------------------------------------- |
| `make build`    | generate templ files, build the web app and the binary |
| `make run`      | build, then run `bin/<name>`                    |
| `make dev`      | run `dev-go` and `dev-web` together             |
| `make generate` | `templ generate`                                |
| `make web`      | `vite build`                                    |
| `make sqlc`     | `sqlc generate`                                 |
| `make migrate`  | `goose up` against `DATABASE_URL`               |
| `make migration name=...` | create the next numbered migration    |
| `make clean`    | remove `bin`, `tmp`, and `web/static`           |

#### Migrations

Migrations live in `internal/db/migrations` and are numbered in order:

```sh
make migration name=add_posts   # writes 00002_add_posts.sql
```

Fill in the `-- +goose Up` and `-- +goose Down` sections, then run
`make migrate` to apply it and `make sqlc` to regenerate the code in
`internal/db`. sqlc reads the migrations as the schema and skips the down
sections. To roll back the last migration, run goose yourself:

```sh
goose -dir internal/db/migrations postgres "$DATABASE_URL" down
```

Run `make sqlc` after changing `queries.sql` too.

## Low level projects

### Prompts

- **Build system**: CMake, or a basic Makefile (C only).
- **Language**: C or C++. Skipped for Makefile projects, which are always C.
- **Library or application**
- **Compiler**: clang or gcc.

The project name is used in CMake, Python, and C identifiers, so it may only
contain letters, numbers, and underscores, and cannot start with a number.

### Requirements

- the compiler you chose
- CMake projects: `cmake` 3.20 or newer and `python3` (gm runs
  `generate_copts.py` once while creating the project). Library projects
  download GoogleTest when you first configure them, so they need network
  access and `git`.
- Makefile projects: `make`

### CMake projects

```
foo/
  CMakeLists.txt             sets the compiler and C/C++ standards
  cmake/helpers.cmake        foo_add_library, foo_add_executable, foo_add_test
  cmake/DownloadGoogleTest.cmake
  foo/CMakeLists.txt         the library and its test, or the executable
  foo/foo.c, foo/foo.h       library sources (foo.cc for C++)
  foo/foo_test.cc            GoogleTest test (libraries only)
  foo/main.c                 application entry point (main.cc for C++)
  foo/copts/                 compile options, see below
  .clang-format
  .gitignore
```

Build and test:

```sh
cmake -B build
cmake --build build
ctest --test-dir build       # libraries
./build/foo/foo              # applications
```

Sources include headers relative to the project root, for example
`#include "foo/foo.h"`. C headers have `extern "C"` guards so they can be used
from C++ tests.

#### Helpers

`cmake/helpers.cmake` defines three functions, each prefixed with the project
name:

```cmake
# compiled library: target foo_bar, alias foo::bar
foo_add_library(
    NAME bar
    HDRS bar.h
    SRCS bar.c
    COPTS ${${PROJECT_NAME}_DEFAULT_COPTS}
    DEPS foo::other
)

# interface (header only) library: leave out SRCS
foo_add_library(
    NAME util
    HDRS util.h
)

# executable: target and binary are named exactly NAME
foo_add_executable(
    NAME tool
    SRCS tool.c
    COPTS ${${PROJECT_NAME}_DEFAULT_COPTS}
    DEPS foo::bar
)

# GoogleTest test: target foo_bar_test; needs include(DownloadGoogleTest)
foo_add_test(
    NAME bar_test
    SRCS bar_test.cc
    COPTS ${${PROJECT_NAME}_TEST_COPTS}
    DEPS foo::bar
)
```

#### Compile options

Warning flags live in `foo/copts/copts.py`, with separate lists for GCC and
Clang:

- `*_FLAGS` apply to both C and C++ files.
- `*_CXX_FLAGS` apply only to C++ files.
- `*_TEST_FLAGS` apply to tests, which are always C++.

After editing `copts.py`, regenerate the CMake file:

```sh
cd foo/copts && python3 generate_copts.py
```

This writes `GENERATED_fooCopts.cmake`, which should not be edited by hand.
`ConfigureCopts.cmake` picks the lists for your compiler and builds
`foo_DEFAULT_COPTS`, wrapping each flag in a `$<COMPILE_LANGUAGE:...>`
generator expression so C files never get C++ only flags. To switch
compilers later, change the two `set(CMAKE_*_COMPILER ...)` lines in the root
`CMakeLists.txt`.

### Makefile projects

```
foo/
  Makefile                   forwards every target to foo/Makefile
  foo/Makefile               the real build
  foo/foo.c, foo/foo.h       library sources
  foo/main.c                 application entry point
  .clang-format
  .gitignore
```

```sh
make            # builds foo/foo or foo/libfoo.a
make clean
```

The root Makefile runs `$(MAKE) -C foo <target>`, so every target and variable
passes through to `foo/Makefile`. That Makefile sets `CC` to the compiler you
chose. `make CC=...` overrides it, and `make OPTIMIZATION=-O0` changes the
optimization level. Makefile projects do not include tests.

## Contributing

### Philosophy

gm is meant to stay small. Before adding something, make sure it solves a
real problem and that it is the simplest way to solve it. We would rather
have fewer features that work well than many that each add a little
complexity. A new prompt or option needs a good reason to exist.

### Working on gm

```sh
make build      # build bin/gm
make test       # go test ./...
make fmt        # go fmt ./...
make clean      # remove bin
```

Run `make fmt` and `make test` before sending changes. When you change what a
project type generates, also run `bin/gm create` and build the generated
project. The tests check which files gm writes and what they contain, but
they do not run create-vite, `go get`, cmake, or the compilers, since those
need network access and extra tools.

### Repository layout

```
main.go             calls cmd.Execute
cmd/
  root.go           the gm command
  create.go         gm create: the prompts for every project type
scaffold/           shared helpers: write templates, run commands, print steps
webapp/
  webapp.go         web app scaffolding
  files/            files written to the project root
  air.toml.tmpl     written only when air is installed
c-cpp/
  ccpp.go           low level scaffolding
  .clang-format     written to every project
  cmake/            root files for CMake projects
  cmake-src/        per language and kind CMakeLists and tests, written to <name>/
  make/             root files for Makefile projects
  src/              per language and kind sources, written to <name>/, used by both
```

The prompts live in `cmd/create.go` and fill an `Options` struct. Each project
package has a `Create(Options)` function that does the work, so the
scaffolding code has no knowledge of the prompts.

### How templates work

Project files are embedded into the gm binary with `go:embed` and written with
`scaffold.Write`:

- Every path is executed as a Go
  [text/template](https://pkg.go.dev/text/template), so a directory named
  `{{.Name}}` becomes the project name.
- Files ending in `.tmpl` are executed as templates and written without the
  `.tmpl` suffix. Other files are copied byte for byte.
- The template data is the project's `Options` struct, so templates can use
  its fields (`{{.Name}}`, `{{.Module}}`) and methods (`{{.CC}}`).
- `{{upper .Name}}` uppercases a value, for example in header guards.

Watch for `${{{` in CMake templates. Go reads the first `{{` as the start of
the action, so `${{{.Name}}_FLAGS}` fails to parse. Use
`${${PROJECT_NAME}_FLAGS}` instead, which CMake expands to the same thing.
The tests catch this, since every template is rendered for every project
variant.

Files that are only written sometimes, like `.air.toml`, live outside the
embedded directory and are written explicitly in the project package.

### Changing a project type

- To change a generated file, edit it under the project type's directory.
  If it now needs a value from the prompts, rename it to end in `.tmpl`.
- To add a file, add it to the right directory. For C and C++ projects, pick
  the directory by when the file is needed (every CMake project, a specific
  language and kind, and so on). Then add it to the expected file list in
  `c-cpp/ccpp_test.go`.
- To add a step that runs a command, add a `scaffold.Step` to the list in
  `Create`. Use `scaffold.Run` so its output is only shown on failure.
- To add a prompt, add a field to `Options`, add a group to the form in
  `cmd/create.go`, and use the field in the templates. Each prompt is its own
  `huh.NewGroup` so gm shows one question at a time.

### Adding a project type

1. Create a package with an embedded `files` directory, an `Options` struct,
   and a `Create(Options) error` function that runs its steps with
   `scaffold.RunSteps` and prints how to start working on the project.
2. Add the type to `ProjectType` in `cmd/create.go`, add it to the project
   type select, and add a function that runs its prompts and calls `Create`.
3. Add tests that write the files for each variant and check the result.
4. Document it in this README and in the description in `cmd/root.go`.

## License

BSD 3-Clause. See [LICENSE](LICENSE).
