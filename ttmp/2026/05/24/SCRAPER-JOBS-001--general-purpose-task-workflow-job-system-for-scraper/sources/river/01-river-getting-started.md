---
Title: River Getting Started
Ticket: SCRAPER-JOBS-001
Status: active
Topics:
    - jobs
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - https://riverqueue.com/docs
Summary: "Captured River documentation source used for backend comparison."
LastUpdated: 2026-05-24T16:24:00-04:00
WhatFor: "Source capture for SCRAPER-JOBS-001 River backend analysis."
WhenToUse: "Use when checking the design document's River claims."
---

Learn how to install River packages for Go, run migrations to get River's database schema in place, and create an initial worker and client to start inserting and working jobs.

[![River Go package docs](https://riverqueue.com/images/badges/go-reference.svg)](https://pkg.go.dev/github.com/riverqueue/river)

---

## Prerequisites

River requires an existing PostgreSQL database, and is most commonly used with [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5). River is tested using the three most recent major versions of PostgreSQL.

## Installation

To install River, run the following in the directory of a Go project (where a `go.mod` file is present):

```sh
go get github.com/riverqueue/river
go get github.com/riverqueue/river/riverdriver/riverpgxv5
```

Alternatively, the `riverdatabasesql` driver can be used instead of `riverpgxv5` for compatibility with Go's built-in `database/sql`. See [inserting jobs with Bun](https://riverqueue.com/docs/bun) or [GORM](https://riverqueue.com/docs/gorm).

## Running migrations

River persists jobs to a Postgres database, and needs a small set of tables created to insert jobs and carry out [leader election](https://riverqueue.com/docs/leader-election). It's bundled with a command line tool which executes migrations, and which future-proofs River in case other migration steps need to be run in future versions.

From the same directory as above, install the River CLI:

```sh
go install github.com/riverqueue/river/cmd/river@latest
```

With the `DATABASE_URL` of a target database (looks like `postgres://host:5432/db`), migrate up:

```sh
river migrate-up --database-url "$DATABASE_URL"
```

See also [migrations](https://riverqueue.com/docs/migrations).

## Job args and workers

Each kind of job in River requires two types: a [`JobArgs`](https://pkg.go.dev/github.com/riverqueue/river#JobArgs) struct and a [`Worker[T JobArgs]`](https://pkg.go.dev/github.com/riverqueue/river#Worker). The `JobArgs` struct has two purposes:

1. It defines the structured arguments for your worker. These arguments are serialized to JSON before the job is stored in the database.
2. It defines a `Kind() string` method that will be used to uniquely identify the kind of job in the database.

Here is a simple `Worker` and `JobArgs` setup for a `SortWorker` which will sort and print a list of strings provided in its arguments:

```go
type SortArgs struct {
    // Strings is a slice of strings to sort.
    Strings []string \`json:"strings"\`
}

func (SortArgs) Kind() string { return "sort" }
```

```go
type SortWorker struct {
    // An embedded WorkerDefaults sets up default methods to fulfill the rest of
    // the Worker interface:
    river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
    sort.Strings(job.Args.Strings)
    fmt.Printf("Sorted strings: %+v\n", job.Args.Strings)
    return nil
}
```

Generics

River utilizes Go generics to simplify your Worker definitions. This means that your worker only needs to deal with fully structured and typed set of arguments. As in the example above, a `Worker` has a 1:1 relationship with the `JobArgs` type it handles.

Jobs are uniquely identified by their "kind" string. Workers are registered on start up so that River knows how to assign jobs to workers:

```go
workers := river.NewWorkers()
// AddWorker panics if the worker is already registered or invalid:
river.AddWorker(workers, &SortWorker{})
```

`AddWorker` panics in case of invalid configuration. Given its succinct syntax and that bad configuration should prevent a worker process from booting, panicking is probably a reasonable compromise for most applications. However, for those who find it distastely, `AddWorkerSafely` is also provided:

```go
workers := river.NewWorkers()
if err := river.AddWorkerSafely(workers, &SortWorker{}); err != nil {
    panic("handle this error")
}
```

## Starting a client

A River [`Client`](https://pkg.go.dev/github.com/riverqueue/river#Client) provides an interface for job insertion and manages job processing and [maintenance services](https://riverqueue.com/docs/maintenance-services). A client is created with a database pool, [driver](https://riverqueue.com/docs/database-drivers), and config struct containing a `Workers` bundle and other settings. Here's a client `Client` working one queue (`"default"`) with up to 100 worker goroutines at a time:

```go
dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
if err != nil {
    // handle error
}

riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        river.QueueDefault: {MaxWorkers: 100},
    },
    Workers: workers,
})
if err != nil {
    // handle error
}

// Run the client inline. All executed jobs will inherit from ctx:
if err := riverClient.Start(ctx); err != nil {
    // handle error
}
```

### Stopping

The client should also be stopped on program shutdown:

```go
// Stop fetching new work and wait for active jobs to finish.
if err := riverClient.Stop(ctx); err != nil {
    // handle error
}
```

There are some complexities around ensuring clients stop cleanly, but also in a timely manner. Read [Graceful shutdown](https://riverqueue.com/docs/graceful-shutdown) for more details on River's stop modes.

[Insert-only clients](https://riverqueue.com/docs/insert-only-clients) will insert jobs, but not work them, and don't need to be started or stopped.

## Inserting jobs

[`Client.InsertTx`](https://pkg.go.dev/github.com/riverqueue/river#Client.InsertTx) is used in conjunction with an instance of job args to insert a job to work on a transaction:

```go
_, err = riverClient.InsertTx(ctx, tx, SortArgs{
    Strings: []string{
        "whale", "tiger", "bear",
    },
}, nil)
if err != nil {
    // handle error
}
```

See the [`InsertAndWork` example](https://pkg.go.dev/github.com/riverqueue/river#example-package-InsertAndWork) for complete code.

[`Client.Insert`](https://pkg.go.dev/github.com/riverqueue/river#Client.Insert) that doesn't take a transaction is also available, although as described in [Transactional enqueuing](https://riverqueue.com/docs/transactional-enqueueing), inserting jobs in transactions is usually more appropriate to avoid bugs.

```go
_, err = riverClient.Insert(ctx, SortArgs{
    Strings: []string{
        "whale", "tiger", "bear",
    },
}, nil)
if err != nil {
    // handle error
}
```

See also [Batch job insertion](https://riverqueue.com/docs/batch-job-insertion).