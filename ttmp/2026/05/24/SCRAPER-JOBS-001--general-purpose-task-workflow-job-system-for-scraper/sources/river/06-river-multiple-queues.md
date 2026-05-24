---
Title: River Multiple Queues
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

River can be configured with any number of queues. All jobs share a single database table regardless of the number of queues in operation, but workers will only select jobs to work for queues that they're configured to handle.

Sample use cases for multiple queues:

- **Guaranteeing worker availability:** A commonly seen pattern is configure a high priority queue with its own set of a workers so that even in the event of a busy general queue, there's always capacity available for priority jobs and they're worked in a timely manner.
- **Sustaining timely throughput:** Similarly, it might be desirable to have a "high effort" queue for jobs that are known to take a long time to execute, like video encoding or LLM training. Keeping expensive jobs in a separate queue helps sustain more timely throughput for other job kinds, which won't be accidentally blocked by long-lived jobs saturating the default queue.
- **Isolating components:** Multiple components/applications that share a single database may all want to use a job queue, but not handle jobs from any other component. River's multiple queues make this easy by naming queues after each component.

---

## Configuring queues

Queues and the maximum number of workers for each are configured through `river.NewClient`:

```go
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        river.QueueDefault: {MaxWorkers: 100},
    },
    Workers: workers,
})
if err != nil {
    // handle error
}
```

Most River examples suggest the use of the default queue, `river.QueueDefault`. There's nothing special about the default queue. All it does is provide a convenient convention that's an appropriate default for most apps, and can be renamed or removed.

More queues are added by putting them in the `Queues` map:

```go
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Logger: slog.New(&slogutil.SlogMessageOnlyHandler{Level: slog.LevelWarn}),
    Queues: map[string]river.QueueConfig{
        river.QueueDefault: {MaxWorkers: 100},
        "high_priority":    {MaxWorkers: 100},
    },
    Workers: workers,
})
```

Real applications will probably want to use a constant instead of a string for queue names (i.e. `"high_priority"`) so that it can be referenced from other parts of the code, like where jobs are inserted.

## Override queue by job kind

Every instance of a job kind can be sent to a specific queue by overriding [`InsertOpts`](https://pkg.go.dev/github.com/riverqueue/river#InsertOpts) on its job args struct, and returning `Queue`:

```go
type AlwaysHighPriorityArgs struct{}

func (AlwaysHighPriorityArgs) Kind() string { return "always_high_priority" }

func (AlwaysHighPriorityArgs) InsertOpts() river.InsertOpts {
  return river.InsertOpts{
    Queue: "high_priority",
  }
}
```

See the [`CustomInsertOpts` example](https://pkg.go.dev/github.com/riverqueue/river#example-package-CustomInsertOpts) for complete code.

## Override queue on insertion

Alternatively, specific job insertions can specify a queue using the `InsertOpts` parameter on `Insert` / `InsertTx`:

```go
_, err = riverClient.Insert(ctx, SometimesHighPriorityArgs{}, &river.InsertOpts{
    Queue: "high_priority",
})
if err != nil {
    // handle error
}
```

## Renaming or removing queues, and compatibility

Like a job's `kind`, its queue is a string that's stored to job records in the database, and renaming or removing a queue might have the unintended consequence of leaving orphaned jobs in the database that no longer have workers that could work them. Jobs with an old queue name may have been inserted while a deploy was going out, or in an error state scheduled for retry.

For safety, renaming or removing a queue should be a two step operation:

1. Rename the queue at all insertion sites, add the new queue name to the River client's `Queues` map (if applicable), but don't remove the old queue. A queue being renamed will have a set of workers configured for both the old name and the new one. Deploy.
2. After observing that all jobs on the old queue have safely completed, remove it from the River client. Deploy again.