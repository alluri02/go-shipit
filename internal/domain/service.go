package domain

import "fmt"

// DeployService is the core domain service — orchestrates deployments.
//
// KEY CONCEPT: Dependency Injection in Go = pass interfaces via constructor.
// No framework. No container. No reflection. Just struct fields + a New function.
//
// C# equivalent (with DI container):
//   public class DeployService {
//       private readonly IDeployRepository _repo;
//       private readonly IQueuePublisher _queue;
//       private readonly INotifier _notifier;
//
//       public DeployService(IDeployRepository repo, IQueuePublisher queue, INotifier notifier) {
//           _repo = repo; _queue = queue; _notifier = notifier;
//       }
//   }
//   // Registered in Startup.cs:
//   services.AddScoped<IDeployService, DeployService>();
//
// Java equivalent (with Spring):
//   @Service
//   public class DeployService {
//       private final DeployRepository repo;
//       private final QueuePublisher queue;
//       private final Notifier notifier;
//
//       @Autowired
//       public DeployService(DeployRepository repo, QueuePublisher queue, Notifier notifier) {
//           this.repo = repo; this.queue = queue; this.notifier = notifier;
//       }
//   }
//
// Go equivalent (NO framework):
//   Just a struct with interface fields + a NewDeployService() function.
//   The caller (main.go) wires everything together.

// DeployService orchestrates the deployment lifecycle.
// It depends on INTERFACES (ports), not concrete implementations.
// This is the Dependency Inversion Principle — the "D" in SOLID.
type DeployService struct {
	repo     DeployRepository
	queue    QueuePublisher
	notifier Notifier
}

// DeployRepository is the interface this service needs for persistence.
// Defined here in the domain (consumer-side) — not in the adapter.
//
// Note: We define interfaces WHERE THEY'RE USED in Go,
// not where they're implemented. This is opposite to C#/Java.
type DeployRepository interface {
	GetByID(id string) (*Deployment, error)
	Save(deployment *Deployment) error
	ListByService(serviceName string, limit int) ([]*Deployment, error)
}

// QueuePublisher is the interface for publishing deploy jobs to a queue.
type QueuePublisher interface {
	Publish(queueName string, message []byte) error
}

// Notifier is the interface for sending notifications.
type Notifier interface {
	Notify(channel, message string) error
}

// NewDeployService creates a DeployService with all dependencies injected.
// This IS the dependency injection — it's just a function call.
//
// C# equivalent: The DI container does this automatically via constructor injection.
// Java equivalent: Spring's @Autowired does this via reflection.
// Go equivalent: You do it manually in main(). Explicit. No magic.
func NewDeployService(repo DeployRepository, queue QueuePublisher, notifier Notifier) *DeployService {
	return &DeployService{
		repo:     repo,
		queue:    queue,
		notifier: notifier,
	}
}

// StartDeploy validates, creates, persists, and enqueues a new deployment.
// Notice how it uses the injected interfaces — never knows the concrete type.
func (s *DeployService) StartDeploy(id, serviceName, imageTag, triggeredBy string, env Environment) (*Deployment, error) {
	// Step 1: Validate input (Lesson 04)
	if err := ValidateDeployment(serviceName, imageTag, triggeredBy); err != nil {
		return nil, fmt.Errorf("StartDeploy: %w", err)
	}

	// Step 2: Create domain object
	deploy := NewDeployment(id, serviceName, imageTag, triggeredBy, env)

	// Step 3: Persist via repository (we don't know if it's MySQL, Postgres, or in-memory)
	if err := s.repo.Save(deploy); err != nil {
		return nil, fmt.Errorf("StartDeploy: save: %w", err)
	}

	// Step 4: Enqueue for processing (we don't know if it's Azure Queue, RabbitMQ, or a channel)
	msg := []byte(fmt.Sprintf(`{"deployment_id":"%s"}`, deploy.ID))
	if err := s.queue.Publish("deployments", msg); err != nil {
		return nil, fmt.Errorf("StartDeploy: enqueue: %w", err)
	}

	// Step 5: Notify (we don't know if it's Slack, email, or stdout)
	_ = s.notifier.Notify("#deploys",
		fmt.Sprintf("🚀 New deploy: %s %s → %s", serviceName, imageTag, env.Name))

	return deploy, nil
}

// GetDeploy retrieves a deployment by ID.
func (s *DeployService) GetDeploy(id string) (*Deployment, error) {
	deploy, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("GetDeploy(%s): %w", id, err)
	}
	return deploy, nil
}

// ListDeploys returns recent deployments for a service.
func (s *DeployService) ListDeploys(serviceName string, limit int) ([]*Deployment, error) {
	deploys, err := s.repo.ListByService(serviceName, limit)
	if err != nil {
		return nil, fmt.Errorf("ListDeploys(%s): %w", serviceName, err)
	}
	return deploys, nil
}
