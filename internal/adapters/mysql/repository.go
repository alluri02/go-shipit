package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/alluri02/go-shipit/internal/domain"
)

// Repository is the MySQL implementation of domain.DeployRepository.
//
// KEY GO CONCEPT: database/sql — Go's standard library for SQL databases.
// It uses raw SQL queries — no ORM (no Entity Framework, no Hibernate).
//
// Why no ORM in Go?
//   1. Go's philosophy: explicit over magic
//   2. Raw SQL is faster (no reflection, no query generation overhead)
//   3. You learn SQL properly (valuable skill)
//   4. Full control over queries (no N+1 surprises)
//
// C# equivalent: You'd typically use Entity Framework Core (ORM)
//   public class DeployRepository : DbContext {
//       public DbSet<Deployment> Deployments { get; set; }
//   }
//   // Or raw with Dapper:
//   var deploy = await conn.QuerySingleAsync<Deployment>("SELECT * FROM ...", new { id });
//
// Java equivalent: You'd typically use JPA/Hibernate (ORM)
//   @Repository
//   public interface DeployRepository extends JpaRepository<Deployment, String> {}
//   // Or raw with JDBI/jOOQ
type Repository struct {
	db *sql.DB
}

// NewRepository creates a MySQL repository.
// The *sql.DB is a connection POOL — safe for concurrent use.
//
// C# equivalent: DbContext (scoped per request) or IDbConnection (pooled)
// Java equivalent: DataSource (connection pool via HikariCP)
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetByID retrieves a deployment by ID.
//
// Key patterns:
//   1. db.QueryRow — for single row queries
//   2. row.Scan — maps columns to struct fields (like Dapper's Query<T>)
//   3. sql.ErrNoRows — sentinel error for "not found"
//
// C# equivalent (Dapper):
//   var deploy = await conn.QuerySingleOrDefaultAsync<Deployment>(
//       "SELECT * FROM deployments WHERE id = @Id", new { Id = id });
//
// Java equivalent (JDBI):
//   return handle.createQuery("SELECT * FROM deployments WHERE id = :id")
//       .bind("id", id)
//       .mapTo(Deployment.class)
//       .findOne();
func (r *Repository) GetByID(id string) (*domain.Deployment, error) {
	query := `
		SELECT id, service_name, image_tag, environment, region, cluster_url,
		       status, risk_score, triggered_by, created_at, updated_at
		FROM deployments
		WHERE id = ?`

	var d domain.Deployment
	var envName, region, clusterURL string
	var status int
	var createdAt, updatedAt time.Time

	// QueryRow executes a query that returns at most one row.
	// Scan copies columns into the provided pointers.
	err := r.db.QueryRow(query, id).Scan(
		&d.ID,
		&d.ServiceName,
		&d.ImageTag,
		&envName,
		&region,
		&clusterURL,
		&status,
		&d.RiskScore,
		&d.TriggeredBy,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.WrapNotFound("deployment", id)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql.GetByID(%s): %w", id, err)
	}

	d.Environment = domain.NewEnvironment(envName, region, clusterURL)
	d.Status = domain.DeployStatus(status)
	d.CreatedAt = createdAt
	d.UpdatedAt = updatedAt

	return &d, nil
}

// Save inserts or updates a deployment (upsert).
//
// Key patterns:
//   1. db.Exec — for INSERT/UPDATE/DELETE (no rows returned)
//   2. ? placeholders — prevents SQL injection (NEVER use fmt.Sprintf for queries!)
//
// C# equivalent (Dapper):
//   await conn.ExecuteAsync(
//       @"INSERT INTO deployments (...) VALUES (@Id, @ServiceName, ...)
//         ON DUPLICATE KEY UPDATE status = @Status, ...",
//       deployment);
//
// Java equivalent (JDBI):
//   handle.createUpdate("INSERT INTO deployments (...) VALUES (:id, :serviceName, ...)")
//       .bindBean(deployment)
//       .execute();
func (r *Repository) Save(deployment *domain.Deployment) error {
	query := `
		INSERT INTO deployments (id, service_name, image_tag, environment, region, 
		                         cluster_url, status, risk_score, triggered_by, 
		                         created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			risk_score = VALUES(risk_score),
			updated_at = VALUES(updated_at)`

	_, err := r.db.Exec(query,
		deployment.ID,
		deployment.ServiceName,
		deployment.ImageTag,
		deployment.Environment.Name,
		deployment.Environment.Region,
		deployment.Environment.ClusterURL,
		int(deployment.Status),
		deployment.RiskScore,
		deployment.TriggeredBy,
		deployment.CreatedAt,
		deployment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("mysql.Save(%s): %w", deployment.ID, err)
	}
	return nil
}

// ListByService retrieves recent deployments for a service.
//
// Key patterns:
//   1. db.Query — for multi-row queries (returns *sql.Rows)
//   2. defer rows.Close() — ALWAYS close rows to release the connection back to pool
//   3. rows.Next() + rows.Scan — iterate row by row
//
// C# equivalent (Dapper):
//   var deploys = await conn.QueryAsync<Deployment>(
//       "SELECT * FROM deployments WHERE service_name = @Name ORDER BY created_at DESC LIMIT @Limit",
//       new { Name = serviceName, Limit = limit });
//
// Java equivalent (JDBI):
//   return handle.createQuery("SELECT * FROM deployments WHERE service_name = :name ...")
//       .bind("name", serviceName)
//       .mapTo(Deployment.class)
//       .list();
func (r *Repository) ListByService(serviceName string, limit int) ([]*domain.Deployment, error) {
	query := `
		SELECT id, service_name, image_tag, environment, region, cluster_url,
		       status, risk_score, triggered_by, created_at, updated_at
		FROM deployments
		WHERE service_name = ?
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, serviceName, limit)
	if err != nil {
		return nil, fmt.Errorf("mysql.ListByService(%s): %w", serviceName, err)
	}
	defer rows.Close() // CRITICAL: always close to return connection to pool

	var results []*domain.Deployment
	for rows.Next() {
		var d domain.Deployment
		var envName, region, clusterURL string
		var status int
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&d.ID, &d.ServiceName, &d.ImageTag,
			&envName, &region, &clusterURL,
			&status, &d.RiskScore, &d.TriggeredBy,
			&createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("mysql.ListByService scan: %w", err)
		}

		d.Environment = domain.NewEnvironment(envName, region, clusterURL)
		d.Status = domain.DeployStatus(status)
		d.CreatedAt = createdAt
		d.UpdatedAt = updatedAt
		results = append(results, &d)
	}

	// Check for iteration errors (network issues mid-read, etc.)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mysql.ListByService rows: %w", err)
	}

	return results, nil
}
