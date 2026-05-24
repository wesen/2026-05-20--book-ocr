---
Title: River Unique Jobs
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

Jobs can be made unique, such that River guarantees that only one exists for a given set of properties. Jobs can be made unique by args, kind, period, queue, and state.

---

## Unique properties

It's occasionally useful to ensure that background work is only performed once for a given set of conditions. For example, there might be a job that does a daily reconcilation of a user's account, but performs heavy lifting across many accounts, so ideally it only runs once a day per account to save on worker resources. River can guarantee job uniqueness along dimensions based on a combination of job properties like arguments and insert period.

Jobs configure unique properties by implementing [`JobArgsWithInsertOpts`](https://pkg.go.dev/github.com/riverqueue/river#JobArgsWithInsertOpts) and populating [`UniqueOpts`](https://pkg.go.dev/github.com/riverqueue/river#UniqueOpts) as part of the [`InsertOpts`](https://pkg.go.dev/github.com/riverqueue/river#InsertOpts) returned, or by adding `UniqueOpts` at insertion time with [`Client.Insert`](https://pkg.go.dev/github.com/riverqueue/river#Client.Insert), `InsertTx`, or any of the `InsertMany*` bulk insertion methods.

```go
type ReconcileAccountArgs struct {
    AccountID int \`json:"account_id"\`
}

func (ReconcileAccountArgs) Kind() string { return "reconcile_account" }

// InsertOpts returns custom insert options that every job of this type will
// inherit, including unique options.
func (ReconcileAccountArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        UniqueOpts: river.UniqueOpts{
            ByArgs:   true,
            ByPeriod: 24 * time.Hour,
        },
    }
}

...

// First job insertion for account 1.
_, err = riverClient.Insert(ctx, ReconcileAccountArgs{AccountID: 1}, nil)
if err != nil {
    panic(err)
}

// Job is inserted a second time, but it doesn't matter because its unique
// args cause the insertion to be skipped because it's meant to only run
// once per account per 24 hour period.
_, err = riverClient.Insert(ctx, ReconcileAccountArgs{AccountID: 1}, nil)
if err != nil {
    panic(err)
}

// Because the job is unique ByArgs, another job for account 2 _is_ allowed.
_, err = riverClient.Insert(ctx, ReconcileAccountArgs{AccountID: 2}, nil)
if err != nil {
    panic(err)
}
```

See the [`UniqueJob` example](https://pkg.go.dev/github.com/riverqueue/river#example-package-UniqueJob) for complete code.

`UniqueOpts` provides options like `ByArgs` and `ByPeriod` that specify the dimensions along which jobs should be considered unique. Each one specified increases the specificity of the unique bound, thereby relaxing the uniqueness of the job (so more options means less uniqueness and more distinct jobs allowed). The job's kind is always taken in account to determine uniqueness, and an empty `UniqueOpts` struct implies no uniqueness. So for example:

- A job with kind `reconcile_account` and `UniqueOpts{ByPeriod: 24 * time.Hour}` means that only one `reconcile_account` can exist for this 24 hour period.
- A job with kind `reconcile_account` and `UniqueOpts{ByPeriod: 24 * time.Hour, ByArgs: true}` means that one `reconcile_account` can exist *per set* of encoded job args and this 24 hour period.
	So, a `reconcile_account` with args `{"account_id":1}` can coexist alongside `{"account_id":2}`.
- A job with kind `reconcile_account` and `UniqueOpts{}` means that no uniqueness is enforced.

### Unique by args

`ByArgs` (taking a boolean) indicates that uniqueness should be enforced by kind and encoded JSON job args. Given args like:

```go
type ReconcileAccountArgs struct {
    AccountID int \`json:"account_id"\`
}
```

The struct `ReconcileAccountArgs{AccountID: 1}` (encoding to `{"account_id":1}`) is allowed to coexist with `ReconcileAccountArgs{AccountID: 2}` (encoding to `{"account_id":2}`), but if another `ReconcileAccountArgs{AccountID: 1}` was inserted, it'd be skipped on grounds of uniqueness.

The keys in the encoded args JSON are sorted alphabetically before being hashed for uniqueness, so the order of keys in the struct doesn't matter. This sorting is *not* recursive, however.

#### Using a subset of args

Sometimes it's desirable to only use a subset of the args for uniqueness. For example you may want only one of a particular kind of job to exist for a given customer, but the args also contain something random like a trace ID. To opt into this mode, the fields considered for uniqueness can be flagged with a struct tag:

```go
type ReconcileAccountArgs struct {
    CustomerID int    \`json:"customer_id" river:"unique"\`
    TraceID    string \`json:"trace_id"\`
}
```

### Unique by period

`ByPeriod` (taking a duration) indicates that uniqueness should be enforced by kind and period. On insertion, the current time is rounded down to the nearest multiple of the given period, and a job is only inserted if there isn't already an existing job that has been created at or scheduled to run between this lower bound and the next multiple of the period.

For example, if a job is inserted with `UniqueOpts{ByPeriod: 15 * time.Minute}` and the current time is 15:21:00, it'll be unique for the interval of 15:15:00 to 15:30:00. A new job inserted at 15:28:00 will be skipped on grounds of uniquess, but one inserted at 15:31:00 would be allowed.

`ByPeriod` is the most commonly used unique property, and other properties are most likely to be specified along with it, rather than be configured by themselves.

### Unique by queue

`ByQueue` (taking a boolean) indicates that uniqueness should be enforced by kind and queue.

For example, if a job with kind `reconcile_account` is inserted into queue `default`, a new insertion of `reconcile_account` would be skipped on grounds of uniqueness, but a `reconcile_account` inserted to queue `high_priority` would be allowed.

### Unique by state

`ByState` (taking a slice of [`JobState`](https://pkg.go.dev/github.com/riverqueue/river/rivertype#JobState)) indicates that uniqueness should be enforced by kind and job state. This is the only unique property that inherits a default if not explicitly assigned, which is all job properties with the exception of `JobStateCancelled` and `JobStateDiscarded`:

```go
[]rivertype.JobState{
    rivertype.JobStateAvailable,
    rivertype.JobStateCompleted,
    rivertype.JobStatePending,
    rivertype.JobStateRunning,
    rivertype.JobStateRetryable,
    rivertype.JobStateScheduled,
}
```

This default is usually the right setting for most unique jobs, but a custom value might be useful in tweaking behavior. For example, removing `JobStateCompleted` from the set above would mean that uniqueness would be enforced within active job states (i.e. being run or available to be run), so that each time a job with this kind completes, a new one is allowed to be enqueued.

Required states

When customizing the `ByState` list, some states are required because River doesn't have conflict resolution for all required internal transitions. The `pending`, `scheduled`, `available`, and `running` states are required whenever customizing this list.

If `JobStateRetryable` is removed from the list, it's possible for an erroring job to hit a conflict when it is retried (because a duplicate has since been inserted). In this scenario, River will move the conflicting job to `discarded` since it cannot be retried.

#### Job retention horizons

When thinking about job state, remember that completed jobs [aren't retained permanently](https://riverqueue.com/docs/maintenance-services#cleaner). The default retention time for completed jobs is 24 hours, so with a default `ByState`, even with no unique period set a new job would be allowed to be inserted every 24 hours as the previous completed job is pruned.

## Checking for skipped inserts

Insert functions return [`JobInsertResult`](https://pkg.go.dev/github.com/riverqueue/river/rivertype#JobInsertResult), containing a `UniqueSkippedAsDuplicate` property that's set to true if an insert was skipped due to uniqueness:

```go
insertRes, err := riverClient.Insert(ctx, SortArgs{...}, nil)
insertRes.UniqueSkippedAsDuplicate // true if job was skipped
```

`JobInsertResult.Job` contains a newly inserted job row, or the preexisting one with matching unique conditions if insertion was skipped.

## At least once

While River can ensure that a unique job is only inserted once, it can't guarantee that it will be worked exactly once. A unique job could work successfully, but fail to have its completed status persisted to the database, which would require that it be worked again for River to be sure it went through. Like other jobs, River provides an at-least-once guarantee for unique jobs.

Unique jobs execute at least once

Although unique jobs ensure that a given job will only be *inserted* once for the chosen properties, those jobs can still execute more than once due to River's at-least-once execution design.

See [Reliable workers](https://riverqueue.com/docs/reliable-workers) for more information.

## Unique index

Job uniqueness is enforced with a special partial unique index on the `river_job` table. The unique index only applies to jobs whose state is within its specified list of unique states, which can be customized on a per-job basis.