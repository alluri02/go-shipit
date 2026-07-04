// Package ports defines the interfaces (contracts) that the domain layer expects.
//
// In hexagonal architecture, "ports" are the boundaries between the domain and the outside world.
// The domain defines WHAT it needs; adapters provide HOW.
//
// KEY GO CONCEPT: Interfaces are implemented IMPLICITLY.
// A type satisfies an interface just by having the right methods — no "implements" keyword.
//
// C# equivalent:  public interface IDeployRepository { ... }
// Java equivalent: public interface DeployRepository { ... }
//
// But in Go, the implementing type NEVER references the interface.
// The compiler checks satisfaction at the point of USE, not at the point of DEFINITION.
package ports

import "github.com/alluri02/go-shipit/internal/domain"

// DeployRepository defines how the domain persists and retrieves deployments.
// Any struct with these methods automatically implements this interface.
//
// C# equivalent:
//   public interface IDeployRepository {
//       Task<Deployment> GetByID(string id);
//       Task Save(Deployment deployment);
//       Task<List<Deployment>> ListByService(string serviceName, int limit);
//   }
//
// Java equivalent:
//   public interface DeployRepository {
//       Optional<Deployment> findById(String id);
//       void save(Deployment deployment);
//       List<Deployment> findByServiceName(String serviceName, int limit);
//   }
type DeployRepository interface {
	GetByID(id string) (*domain.Deployment, error)
	Save(deployment *domain.Deployment) error
	ListByService(serviceName string, limit int) ([]*domain.Deployment, error)
}

// QueuePublisher defines how the domain publishes messages to a queue.
// The webhookreceiver uses this to enqueue jobs; the processor consumes them.
//
// C# equivalent:
//   public interface IQueuePublisher {
//       Task Publish(string queueName, byte[] message);
//   }
//
// Java equivalent:
//   public interface QueuePublisher {
//       void publish(String queueName, byte[] message);
//   }
type QueuePublisher interface {
	Publish(queueName string, message []byte) error
}

// QueueConsumer defines how the domain reads messages from a queue.
// Returns a channel of messages — a Go-native way to stream data (Lesson 07).
//
// C# equivalent:
//   public interface IQueueConsumer {
//       IAsyncEnumerable<QueueMessage> Consume(string queueName, CancellationToken ct);
//   }
//
// Java equivalent:
//   public interface QueueConsumer {
//       Flux<QueueMessage> consume(String queueName); // Reactor
//   }
type QueueConsumer interface {
	Consume(queueName string) (<-chan []byte, error)
}

// ImageBuilder defines how the domain builds and pushes container images.
//
// C# equivalent:
//   public interface IImageBuilder {
//       Task<string> Build(string serviceName, string tag, string dockerfilePath);
//       Task Push(string imageRef);
//   }
//
// Java equivalent:
//   public interface ImageBuilder {
//       String build(String serviceName, String tag, String dockerfilePath);
//       void push(String imageRef);
//   }
type ImageBuilder interface {
	Build(serviceName, tag, dockerfilePath string) (imageRef string, err error)
	Push(imageRef string) error
}

// Deployer defines how the domain deploys an image to a target environment.
//
// C# equivalent:
//   public interface IDeployer {
//       Task<DeployResult> Deploy(string imageRef, Environment env);
//   }
//
// Java equivalent:
//   public interface Deployer {
//       DeployResult deploy(String imageRef, Environment env);
//   }
type Deployer interface {
	Deploy(imageRef string, env domain.Environment) error
	Rollback(deploymentID string) error
}

// Notifier defines how the domain sends notifications (Slack, email, etc.).
//
// C# equivalent:
//   public interface INotifier {
//       Task Notify(string channel, string message);
//   }
//
// Java equivalent:
//   public interface Notifier {
//       void notify(String channel, String message);
//   }
type Notifier interface {
	Notify(channel, message string) error
}

// RiskAnalyzer defines how the domain gets AI-powered risk assessments.
//
// C# equivalent:
//   public interface IRiskAnalyzer {
//       Task<RiskAssessment> Assess(DeployContext context);
//   }
//
// Java equivalent:
//   public interface RiskAnalyzer {
//       RiskAssessment assess(DeployContext context);
//   }
type RiskAnalyzer interface {
	Assess(deployment *domain.Deployment, diffSummary string) (score int, reason string, err error)
}
