package domain_test

import (
	"errors"
	"testing"

	"github.com/alluri02/go-shipit/internal/domain"
)

// --- Mock Implementations ---
// In Go, mocking is free — just create a struct with the right methods.
// No Moq (C#), no Mockito (Java), no code generation needed.
//
// C# equivalent:
//   var mockRepo = new Mock<IDeployRepository>();
//   mockRepo.Setup(r => r.Save(It.IsAny<Deployment>())).Returns(Task.CompletedTask);
//
// Java equivalent:
//   DeployRepository mockRepo = mock(DeployRepository.class);
//   when(mockRepo.save(any())).thenReturn(null);

type mockRepo struct {
	deployments map[string]*domain.Deployment
	saveErr     error // Set this to simulate save failures
}

func newMockRepo() *mockRepo {
	return &mockRepo{deployments: make(map[string]*domain.Deployment)}
}

func (m *mockRepo) GetByID(id string) (*domain.Deployment, error) {
	d, ok := m.deployments[id]
	if !ok {
		return nil, domain.WrapNotFound("deployment", id)
	}
	return d, nil
}

func (m *mockRepo) Save(d *domain.Deployment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.deployments[d.ID] = d
	return nil
}

func (m *mockRepo) ListByService(name string, limit int) ([]*domain.Deployment, error) {
	var results []*domain.Deployment
	for _, d := range m.deployments {
		if d.ServiceName == name {
			results = append(results, d)
		}
	}
	return results, nil
}

type mockQueue struct {
	messages [][]byte
	err      error
}

func (m *mockQueue) Publish(queueName string, message []byte) error {
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, message)
	return nil
}

type mockNotifier struct {
	notifications []string
}

func (m *mockNotifier) Notify(channel, message string) error {
	m.notifications = append(m.notifications, channel+": "+message)
	return nil
}

// --- Service Tests ---

func TestDeployService_StartDeploy(t *testing.T) {
	repo := newMockRepo()
	queue := &mockQueue{}
	notifier := &mockNotifier{}
	service := domain.NewDeployService(repo, queue, notifier)

	env := domain.NewEnvironment("staging", "eastus", "")
	deploy, err := service.StartDeploy("d-001", "payments-api", "v2.4.1", "api", env)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deploy.ID != "d-001" {
		t.Errorf("ID = %q, want %q", deploy.ID, "d-001")
	}
	if deploy.Status != domain.DeployStatusPending {
		t.Errorf("Status = %v, want Pending", deploy.Status)
	}

	// Verify side effects
	if len(queue.messages) != 1 {
		t.Errorf("expected 1 queued message, got %d", len(queue.messages))
	}
	if len(notifier.notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(notifier.notifications))
	}

	// Verify persisted
	saved, err := repo.GetByID("d-001")
	if err != nil {
		t.Fatalf("deployment not saved: %v", err)
	}
	if saved.ServiceName != "payments-api" {
		t.Errorf("saved ServiceName = %q, want %q", saved.ServiceName, "payments-api")
	}
}

func TestDeployService_StartDeploy_ValidationError(t *testing.T) {
	repo := newMockRepo()
	queue := &mockQueue{}
	notifier := &mockNotifier{}
	service := domain.NewDeployService(repo, queue, notifier)

	env := domain.NewEnvironment("staging", "eastus", "")
	_, err := service.StartDeploy("d-001", "", "v1.0", "api", env) // empty service name

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var valErr *domain.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestDeployService_StartDeploy_SaveError(t *testing.T) {
	repo := newMockRepo()
	repo.saveErr = errors.New("database connection lost")
	queue := &mockQueue{}
	notifier := &mockNotifier{}
	service := domain.NewDeployService(repo, queue, notifier)

	env := domain.NewEnvironment("staging", "eastus", "")
	_, err := service.StartDeploy("d-001", "svc", "v1.0", "api", env)

	if err == nil {
		t.Fatal("expected save error, got nil")
	}
	if len(queue.messages) != 0 {
		t.Error("should not enqueue if save fails")
	}
}

func TestDeployService_GetDeploy_NotFound(t *testing.T) {
	repo := newMockRepo()
	queue := &mockQueue{}
	notifier := &mockNotifier{}
	service := domain.NewDeployService(repo, queue, notifier)

	_, err := service.GetDeploy("nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
