---
Title: River Scheduled Jobs
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

Jobs can be scheduled to run at a future time and date instead of running immediately.

---

## Basic usage

At insertion time, any job can specify a `ScheduledAt` as part of its `InsertOpts` to run it at a future time. The following code inserts a job that will run three hours from now:

```go
_, err = riverClient.Insert(ctx,
    ScheduledArgs{
        Message: "hello from the future",
    },
    &river.InsertOpts{
        ScheduledAt: time.Now().Add(3 * time.Hour),
    }
)
if err != nil {
    // handle error
}
```

See the [`ScheduledJob` example](https://pkg.go.dev/github.com/riverqueue/river#example-package-ScheduledJob) for complete code.

This job will be inserted into the queue with a `scheduled` state and the specified `scheduled_at` time. Once that time has elapsed, the next loop of the [Scheduler](https://riverqueue.com/docs/maintenance-services#scheduler) will move it to `available` so it can be picked up by an available Client. This means there will always be some delay after the scheduled time (generally less than 5 seconds), so this option is not suitable for running jobs only a few seconds in the future unless the added delay is acceptable.