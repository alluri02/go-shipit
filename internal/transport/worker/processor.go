package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// Processor is the webhook processor — consumes jobs from a channel and processes them.
//
// KEY GO CONCEPT: Goroutines + Channels
//
// A goroutine is a lightweight thread managed by the Go runtime.
// A channel is a typed pipe for communication between goroutines.
//
// Together they implement CSP (Communicating Sequential Processes):
//   "Don't communicate by sharing memory; share memory by communicating."
//
// C# equivalent: Task.Run() + Channel<T> (System.Threading.Channels)
//   var channel = Channel.CreateUnbounded<Job>();
//   for (int i = 0; i < workerCount; i++)
//       _ = Task.Run(() => ProcessLoop(channel.Reader));
//
// Java equivalent: ExecutorService + BlockingQueue<Job>
//   ExecutorService pool = Executors.newFixedThreadPool(workerCount);
//   BlockingQueue<Job> queue = new LinkedBlockingQueue<>();
//   for (int i = 0; i < workerCount; i++)
//       pool.submit(() -> processLoop(queue));
type Processor struct {
	service    *domain.DeployService
	jobs       chan DeployJob       // Channel — the pipe between producer and consumer
	numWorkers int
	wg         sync.WaitGroup      // WaitGroup — wait for all goroutines to finish
}

// DeployJob represents a unit of work to be processed.
type DeployJob struct {
	DeploymentID string `json:"deployment_id"`
	ServiceName  string `json:"service_name"`
	ImageTag     string `json:"image_tag"`
	Environment  string `json:"environment"`
	Region       string `json:"region"`
}

// NewProcessor creates a processor with a buffered channel and worker count.
//
// Buffered channel = channel with capacity. Producers can send without blocking
// until the buffer is full. Like a bounded queue.
//
// C# equivalent: Channel.CreateBounded<Job>(new BoundedChannelOptions(100));
// Java equivalent: new ArrayBlockingQueue<>(100);
func NewProcessor(service *domain.DeployService, numWorkers, bufferSize int) *Processor {
	return &Processor{
		service:    service,
		jobs:       make(chan DeployJob, bufferSize), // buffered channel
		numWorkers: numWorkers,
	}
}

// Start launches worker goroutines. Each one reads from the jobs channel.
//
// KEY INSIGHT: `go func()` launches a goroutine — a lightweight "thread".
// Goroutines cost ~2KB of stack (vs ~1MB for OS threads in C#/Java).
// You can run millions of them.
//
// C# equivalent:
//   for (int i = 0; i < workerCount; i++)
//       _ = Task.Run(async () => { await foreach (var job in channel.Reader.ReadAllAsync()) ... });
//
// Java equivalent:
//   for (int i = 0; i < workerCount; i++)
//       executor.submit(() -> { while (true) { Job job = queue.take(); process(job); } });
func (p *Processor) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)           // Track this goroutine
		go p.worker(i)        // Launch goroutine — `go` keyword is all it takes!
	}
	log.Printf("Processor started: %d workers, buffer size %d", p.numWorkers, cap(p.jobs))
}

// Submit sends a job to the processor. Non-blocking if buffer has space.
//
// The `<-` operator sends to or receives from a channel:
//   ch <- value    // send
//   value := <-ch  // receive
//
// C# equivalent: await channel.Writer.WriteAsync(job);
// Java equivalent: queue.put(job);
func (p *Processor) Submit(job DeployJob) {
	p.jobs <- job // Send job into the channel
}

// SubmitJSON parses a JSON message and submits it as a job.
func (p *Processor) SubmitJSON(message []byte) error {
	var job DeployJob
	if err := json.Unmarshal(message, &job); err != nil {
		return fmt.Errorf("invalid job JSON: %w", err)
	}
	p.Submit(job)
	return nil
}

// Stop signals workers to stop and waits for them to finish.
//
// Closing a channel signals all receivers that no more values will be sent.
// The `for job := range ch` loop exits when the channel is closed.
//
// C# equivalent: channel.Writer.Complete(); await Task.WhenAll(workers);
// Java equivalent: executor.shutdown(); executor.awaitTermination(...);
func (p *Processor) Stop() {
	close(p.jobs)  // Close channel — workers will exit their range loops
	p.wg.Wait()   // Wait for all workers to finish current jobs
	log.Println("Processor stopped: all workers finished")
}

// worker is the goroutine function — reads jobs from the channel until it's closed.
//
// `for job := range p.jobs` — this is Go's idiomatic way to consume from a channel.
// It blocks waiting for the next value, and exits when the channel is closed.
//
// C# equivalent:
//   await foreach (var job in channel.Reader.ReadAllAsync(cancellationToken)) { ... }
//
// Java equivalent:
//   while (!Thread.currentThread().isInterrupted()) {
//       Job job = queue.take();  // blocks until available
//       process(job);
//   }
func (p *Processor) worker(id int) {
	defer p.wg.Done() // Signal completion when this goroutine exits

	log.Printf("  Worker %d: started", id)

	for job := range p.jobs { // Blocks until a job arrives or channel is closed
		p.processJob(id, job)
	}

	log.Printf("  Worker %d: stopped (channel closed)", id)
}

// processJob handles a single deployment job.
func (p *Processor) processJob(workerID int, job DeployJob) {
	log.Printf("  Worker %d: processing %s (%s %s → %s)",
		workerID, job.DeploymentID, job.ServiceName, job.ImageTag, job.Environment)

	// Simulate build + push + deploy stages
	stages := []struct {
		name   string
		status domain.DeployStatus
		delay  time.Duration
	}{
		{"building", domain.DeployStatusBuilding, 500 * time.Millisecond},
		{"pushing", domain.DeployStatusPushing, 300 * time.Millisecond},
		{"deploying", domain.DeployStatusDeploying, 400 * time.Millisecond},
		{"succeeded", domain.DeployStatusSucceeded, 0},
	}

	for _, stage := range stages {
		if stage.delay > 0 {
			time.Sleep(stage.delay) // Simulate work
		}
		log.Printf("  Worker %d: [%s] %s → %s",
			workerID, job.DeploymentID, stage.name, stage.status)
	}

	log.Printf("  Worker %d: ✓ completed %s", workerID, job.DeploymentID)
}
