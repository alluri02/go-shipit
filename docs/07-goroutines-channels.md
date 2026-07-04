# Lesson 07: Goroutines & Channels

## What We Built
```
go-shipit/
└── internal/
    └── transport/
        └── worker/
            └── processor.go    ← Worker pool: goroutines + channels
```

---

## The Core Difference

| | Go | C# | Java |
|-|-----|-----|------|
| **Lightweight thread** | `go func()` (goroutine) | `Task.Run()` | `executor.submit()` |
| **Cost per unit** | ~2KB stack | ~1MB (thread pool) | ~1MB (thread) |
| **Communication** | Channels (`chan T`) | `Channel<T>` | `BlockingQueue<T>` |
| **Max concurrent** | Millions | Thousands (thread pool) | Thousands (OS threads) |
| **Philosophy** | "Share memory by communicating" | Shared memory + locks | Shared memory + locks |

---

## Goroutines: Lightweight "Threads"

A goroutine is launched with the `go` keyword:

```go
go doWork()           // Launch goroutine — that's it!
go func() {          // Anonymous goroutine
    fmt.Println("running concurrently")
}()
```

### C# Equivalent:
```csharp
_ = Task.Run(() => DoWork());
// or
_ = Task.Run(async () => {
    Console.WriteLine("running concurrently");
});
```

### Java Equivalent:
```java
executor.submit(() -> doWork());
// or
CompletableFuture.runAsync(() -> {
    System.out.println("running concurrently");
});
```

### Why goroutines are different:
- **2KB initial stack** (grows as needed) vs **1MB fixed stack** for OS threads
- **Multiplexed onto OS threads** by Go's scheduler (M:N threading)
- **No thread pool configuration** — just `go` and forget
- **Can run millions** on a single machine

---

## Channels: Typed Pipes Between Goroutines

A channel is a typed conduit for sending values between goroutines:

```go
// Create a channel of strings
ch := make(chan string)

// Send a value INTO the channel
ch <- "hello"

// Receive a value FROM the channel
msg := <-ch
```

### Buffered vs Unbuffered:

```go
// Unbuffered — sender blocks until receiver is ready (synchronization point)
ch := make(chan int)

// Buffered — sender blocks only when buffer is full (like a queue)
ch := make(chan int, 100)
```

### C# Equivalent:
```csharp
// System.Threading.Channels
var ch = Channel.CreateUnbounded<string>();

// Send
await ch.Writer.WriteAsync("hello");

// Receive
var msg = await ch.Reader.ReadAsync();
```

### Java Equivalent:
```java
// java.util.concurrent
BlockingQueue<String> ch = new LinkedBlockingQueue<>();

// Send
ch.put("hello");

// Receive
String msg = ch.take();
```

---

## The Worker Pool Pattern

This is the most common concurrency pattern in Go:

```
Producer → [Channel (buffer)] → Worker 1 → process
                               → Worker 2 → process
                               → Worker 3 → process
```

```go
func main() {
    jobs := make(chan Job, 100)   // Buffered channel = job queue

    // Launch workers
    for i := 0; i < 3; i++ {
        go worker(i, jobs)
    }

    // Submit jobs
    for _, j := range jobList {
        jobs <- j
    }

    close(jobs)  // Signal: no more jobs coming
}

func worker(id int, jobs <-chan Job) {  // <-chan = receive-only channel
    for job := range jobs {             // range exits when channel is closed
        process(job)
    }
}
```

### C# Equivalent:
```csharp
var channel = Channel.CreateBounded<Job>(100);

// Workers
var workers = Enumerable.Range(0, 3).Select(i =>
    Task.Run(async () => {
        await foreach (var job in channel.Reader.ReadAllAsync()) {
            await ProcessAsync(job);
        }
    })
).ToArray();

// Submit
foreach (var job in jobList)
    await channel.Writer.WriteAsync(job);

channel.Writer.Complete();
await Task.WhenAll(workers);
```

### Java Equivalent:
```java
BlockingQueue<Job> queue = new ArrayBlockingQueue<>(100);
ExecutorService pool = Executors.newFixedThreadPool(3);

// Workers
for (int i = 0; i < 3; i++) {
    pool.submit(() -> {
        while (true) {
            Job job = queue.poll(1, TimeUnit.SECONDS);
            if (job == null) break;
            process(job);
        }
    });
}

// Submit
for (Job job : jobList) queue.put(job);
pool.shutdown();
pool.awaitTermination(1, TimeUnit.MINUTES);
```

---

## Deep Dive: Channel Direction

Go lets you restrict channel direction in function signatures:

```go
func producer(out chan<- int) {   // can ONLY send to out
    out <- 42
}

func consumer(in <-chan int) {    // can ONLY receive from in
    val := <-in
}
```

This is compile-time enforced — prevents bugs where a consumer accidentally sends.

### No C#/Java equivalent — direction is just a convention in those languages.

---

## Deep Dive: sync.WaitGroup

`WaitGroup` waits for a collection of goroutines to finish:

```go
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
    wg.Add(1)         // "I'm starting one goroutine"
    go func(id int) {
        defer wg.Done()  // "This goroutine is done"
        doWork(id)
    }(i)
}

wg.Wait()  // Block until all goroutines call Done()
```

### C# Equivalent:
```csharp
var tasks = new List<Task>();
for (int i = 0; i < 3; i++) {
    tasks.Add(Task.Run(() => DoWork(i)));
}
await Task.WhenAll(tasks);
```

### Java Equivalent:
```java
CountDownLatch latch = new CountDownLatch(3);
for (int i = 0; i < 3; i++) {
    executor.submit(() -> { doWork(); latch.countDown(); });
}
latch.await();
```

---

## Deep Dive: `range` Over Channels

```go
for job := range jobs {
    process(job)
}
// This loop:
// 1. Blocks waiting for next value
// 2. Processes it
// 3. Repeats
// 4. Exits when channel is closed AND empty
```

### C# Equivalent:
```csharp
await foreach (var job in channel.Reader.ReadAllAsync()) {
    Process(job);
}
```

### Java Equivalent:
No direct equivalent — you typically use `poll()` with a sentinel or poison pill.

---

## Deep Dive: `close()` — Signaling Completion

```go
close(jobs)  // Tells all receivers: "no more values will be sent"
```

After close:
- **Sending** → panic
- **Receiving** → returns remaining buffered values, then zero value + `false`
- **`range`** → exits

### C# Equivalent:
```csharp
channel.Writer.Complete();
```

### Java Equivalent:
No direct equivalent — typically use a "poison pill" (special sentinel value).

---

## How This Fits in ShipIt

```
GitHub Webhook arrives
       ↓
webhookreceiver validates + enqueues
       ↓
[Azure Queue / Channel buffer]
       ↓
webhookprocessor workers (goroutines)
  - Worker 0: build → push → deploy
  - Worker 1: build → push → deploy
  - Worker 2: build → push → deploy
```

Each worker processes one deployment at a time. With 3 workers, we process 3 deployments concurrently.

---

## Try It

```bash
go build -o shipit.exe ./cmd/shipit
.\shipit.exe process
```

You'll see 5 jobs processed by 3 workers concurrently — notice the interleaved output!

---

## Key Takeaways

1. **`go func()`** — launches a goroutine. Costs ~2KB. Run millions of them.
2. **`make(chan T, size)`** — creates a channel. The pipe between goroutines.
3. **`ch <- val` / `val := <-ch`** — send/receive. Blocks when appropriate.
4. **`for x := range ch`** — consume until channel is closed.
5. **`close(ch)`** — signal "no more values" to all receivers.
6. **`sync.WaitGroup`** — wait for a group of goroutines to finish.
7. **Worker pool** = goroutines + buffered channel. The #1 Go pattern.

---

## Next: [Lesson 08 — Testing (Table-Driven)](./08-testing.md)
We'll write unit tests for the domain layer using Go's built-in testing framework.
