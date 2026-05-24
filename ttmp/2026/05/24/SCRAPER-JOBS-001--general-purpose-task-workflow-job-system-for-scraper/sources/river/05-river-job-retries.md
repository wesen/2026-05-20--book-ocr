---
Title: River Job Retries
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

When River jobs encounter an error or other failure, they are automatically retried after a delay. The default retry policy backs off for `attempts ^ 4 + rand(±10%)` seconds (0s, 1s, 16s,...), and jobs will be tried a maximum of 25 times by default.

---

## Limiting retry attempts

River's default is to retry jobs up to 25 times as defined by `river.DefaultMaxAttempts`. This can be customized for a given worker using the optional `JobArgsWithInsertOpts` interface on the `JobArgs` implementation:

```go
type RetryOnceJobArgs struct {}

func (args RetryOnceJobArgs) Kind() string { return "RetryOnceJob" }

func (args RetryOnceJobArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{MaxAttempts: 1}
}
```

The max attempts can also be customized for an individual job at insertion time. This takes precedence over a job-level default:

```go
_, err = riverClient.Insert(ctx, RetryOnceJobArgs{}, &river.InsertOpts{
    // use a max attempts of 5 for this one job even though the job has a
    // default limit of 1:
    MaxAttempts: 5,
})
```

---

## Retry delays

Jobs are typically delayed for some amount of time between attempts. River provides reasonable default retry delays, but this behavior is also fully customizable at the client and worker level.

### Client retry policy

At the client level, River provides a configurable `RetryPolicy` option which defaults to `DefaultClientRetryPolicy`. The default retry policy uses an exponential backoff based on how many times the job has errored. A randomized ±10% jitter is applied to prevent stampeding errors from all retrying at the same time.

```ruby
attempts ^ 4 + rand(±10%)
```

A job which has errored repeatedly will see its retries delayed as shown below:

| Attempt | Delay (±10%) |  |
| --- | --- | --- |
| 0 → 1 | \- | Initial run before an error (no delay) |
| 1 → 2 | 1s | Delay after first error |
| 2 → 3 | 16s |  |
| 3 → 4 | 1m21s |  |
| 4 → 5 | 4m16s |  |
| 5 → 6 | 10m25s |  |
| ... | ... |  |
| 24 → 25 | 3d20h9m36s | Last retry, ~3 weeks after first run |

The last retry comes about **three weeks** after the first time a job is worked, so in case of a buggy job, there's plenty of time to get a fix out before the job is finally discarded.

The client retry policy can easily be customized:

```go
// LinearRetryPolicy delays subsequent retries by 5 seconds for each time
// the job has failed (5s, 10s, 15s, etc.).
type LinearRetryPolicy struct {}

// NextRetry returns the next retry time based on the non-generic JobRow
// which includes an up-to-date Errors list.
func (policy *LinearRetryPolicy) NextRetry(job *rivertype.JobRow) time.Time {
    // The latest error is not yet included in the job's Errors list, so we
    // add 1 to the length to account for that.
    return time.Now().Add((len(job.Errors) + 1) * 5 * time.Second)
}

client, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    RetryPolicy: &LinearRetryPolicy{},
    // ...
})
```

### Customizing retry delays for a specific worker

While the client retry policy applies to all kinds of jobs worked by that client, workers can also override the retry behavior for all jobs of that kind. This is done via the `Worker` interface's `NextRetry(*Job[T]) time.Time` method. `WorkerDefaults` always returns 0 for this method to fallback to the client retry policy, so it can be customized for a given worker just by returning a non-zero value:

```go
type ConstantRetryTimeWorker struct {
    river.WorkerDefaults[MyJobArgs]
}

func (w *ConstantRetryTimeWorker) Work(job *Job[MyJobArgs]) error {
    // ...
}

// NextRetry always schedules the next retry for 10 seconds from now.
func (w *ConstantRetryTimeWorker) NextRetry(job *Job[MyJobArgs]) time.Time {
    return time.Now().Add(10*time.Second)
}
```