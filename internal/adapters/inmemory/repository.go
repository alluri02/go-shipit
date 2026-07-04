package inmemory

import (
	"sync"

	"github.com/alluri02/go-shipit/internal/domain"
)

// Repository is an in-memory implementation of domain.DeployRepository.
// Used for local development and testing — no database needed.
//
// This struct satisfies the DeployRepository interface IMPLICITLY (Lesson 03).
// It never imports or references the interface — it just has the right methods.
//
// C# equivalent:
//   public class InMemoryDeployRepository : IDeployRepository { ... }
//
// Java equivalent:
//   public class InMemoryDeployRepository implements DeployRepository { ... }
type Repository struct {
	mu          sync.RWMutex           // Protects concurrent access (goroutine-safe)
	deployments map[string]*domain.Deployment
}

// NewRepository creates an empty in-memory repository.
func NewRepository() *Repository {
	return &Repository{
		deployments: make(map[string]*domain.Deployment),
	}
}

// GetByID retrieves a deployment by ID. Returns ErrNotFound if not present.
func (r *Repository) GetByID(id string) (*domain.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.deployments[id]
	if !ok {
		return nil, domain.WrapNotFound("deployment", id)
	}
	return d, nil
}

// Save stores a deployment. Overwrites if already exists.
func (r *Repository) Save(deployment *domain.Deployment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.deployments[deployment.ID] = deployment
	return nil
}

// ListByService returns deployments for a given service name, up to limit.
func (r *Repository) ListByService(serviceName string, limit int) ([]*domain.Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.Deployment
	for _, d := range r.deployments {
		if d.ServiceName == serviceName {
			results = append(results, d)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}
