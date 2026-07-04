# Lesson 10: Database (SQL, No ORM)

## What We Built
```
go-shipit/
├── internal/
│   └── adapters/
│       └── mysql/
│           ├── connect.go        ← Connection pool setup
│           └── repository.go     ← DeployRepository implementation (raw SQL)
└── schema/
    ├── .skeema                   ← Skeema config (per-environment)
    └── deployments.sql           ← Table definition
```

---

## The Core Difference

| | Go (`database/sql`) | C# (Entity Framework) | Java (JPA/Hibernate) |
|-|---------------------|----------------------|---------------------|
| **Approach** | Raw SQL | ORM (LINQ to SQL) | ORM (HQL/JPQL) |
| **Mapping** | Manual `Scan()` | Auto (DbSet<T>) | Auto (@Entity) |
| **Migrations** | Skeema (schema diff) | EF Migrations | Flyway / Liquibase |
| **Connection pool** | Built-in (`sql.DB`) | Built-in (EF context pooling) | HikariCP |
| **Driver model** | Blank import `_ "driver"` | NuGet package | JDBC ServiceLoader |
| **N+1 problem** | Impossible (you write the SQL) | Common trap | Common trap |

---

## Why No ORM in Go?

Go community strongly favors raw SQL. Here's why:

1. **Explicit > Magic** — you see exactly what query runs
2. **Performance** — no reflection, no query generation overhead
3. **No surprises** — no lazy loading, no N+1, no unexpected JOINs
4. **SQL is a skill** — ORMs hide it; Go embraces it
5. **Simplicity** — `database/sql` is ~10 functions total

### The tradeoff:
You write more code. But you never debug mysterious ORM-generated queries.

---

## Pattern 1: Connection Pool

```go
db, err := sql.Open("mysql", dsn)  // Does NOT connect yet!
db.SetMaxOpenConns(25)              // Max simultaneous connections
db.SetMaxIdleConns(5)               // Keep 5 ready
db.SetConnMaxLifetime(5 * time.Minute)
db.Ping()                           // NOW it actually connects
```

### C# Equivalent:
```csharp
// EF Core — pooling is automatic
services.AddDbContextPool<AppDbContext>(options =>
    options.UseMySql(connectionString, ServerVersion.AutoDetect(connectionString)));
```

### Java Equivalent:
```java
// HikariCP — most popular connection pool
HikariConfig config = new HikariConfig();
config.setJdbcUrl("jdbc:mysql://localhost/shipit");
config.setMaximumPoolSize(25);
config.setMinimumIdle(5);
DataSource ds = new HikariDataSource(config);
```

---

## Pattern 2: QueryRow + Scan (Single Row)

```go
var name string
var age int
err := db.QueryRow("SELECT name, age FROM users WHERE id = ?", id).Scan(&name, &age)

if err == sql.ErrNoRows {
    // Not found — map to domain error
    return nil, ErrNotFound
}
if err != nil {
    return nil, err
}
```

### Key points:
- `?` = placeholder (prevents SQL injection — NEVER use string concatenation!)
- `.Scan(&field)` = maps columns to variables by position
- `sql.ErrNoRows` = sentinel error for "no results"

### C# Equivalent (Dapper):
```csharp
var user = await conn.QuerySingleOrDefaultAsync<User>(
    "SELECT name, age FROM users WHERE id = @Id", new { Id = id });
if (user == null) throw new NotFoundException();
```

### Java Equivalent (JDBI):
```java
Optional<User> user = handle.createQuery("SELECT name, age FROM users WHERE id = :id")
    .bind("id", id)
    .mapTo(User.class)
    .findOne();
```

---

## Pattern 3: Query + Rows (Multiple Rows)

```go
rows, err := db.Query("SELECT id, name FROM users WHERE active = ?", true)
if err != nil { return nil, err }
defer rows.Close()  // ← CRITICAL: releases connection back to pool

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return nil, err
    }
    users = append(users, u)
}
if err := rows.Err(); err != nil {  // Check for iteration errors
    return nil, err
}
return users, nil
```

### IMPORTANT: Always `defer rows.Close()`
If you forget, the connection is leaked and your pool eventually exhausts.

### C# Equivalent (Dapper):
```csharp
var users = await conn.QueryAsync<User>(
    "SELECT id, name FROM users WHERE active = @Active", new { Active = true });
// Dapper handles rows lifecycle automatically
```

### Java Equivalent:
```java
List<User> users = handle.createQuery("SELECT id, name FROM users WHERE active = :active")
    .bind("active", true)
    .mapTo(User.class)
    .list();
// JDBI handles ResultSet lifecycle
```

---

## Pattern 4: Exec (INSERT/UPDATE/DELETE)

```go
result, err := db.Exec(
    "INSERT INTO users (id, name) VALUES (?, ?)",
    user.ID, user.Name,
)
if err != nil { return err }

rowsAffected, _ := result.RowsAffected()
fmt.Printf("Inserted %d rows\n", rowsAffected)
```

### C# Equivalent (Dapper):
```csharp
var affected = await conn.ExecuteAsync(
    "INSERT INTO users (id, name) VALUES (@Id, @Name)", user);
```

### Java Equivalent:
```java
int affected = handle.createUpdate("INSERT INTO users (id, name) VALUES (:id, :name)")
    .bindBean(user)
    .execute();
```

---

## Pattern 5: SQL Injection Prevention

```go
// ✓ SAFE — parameterized query
db.QueryRow("SELECT * FROM users WHERE id = ?", userInput)

// ✗ DANGEROUS — SQL injection vulnerability!
db.QueryRow("SELECT * FROM users WHERE id = '" + userInput + "'")
// If userInput = "'; DROP TABLE users; --" → your table is gone
```

### The rule: ALWAYS use `?` placeholders. Never concatenate user input into SQL.

This is the same in all languages:
- Go: `?`
- C#: `@ParamName`
- Java: `:paramName` or `?`

---

## Schema Management: Skeema

Skeema is a CLI tool that diffs your `.sql` files against the live database:

```bash
# See what would change (like terraform plan)
skeema diff development

# Apply changes (like terraform apply)
skeema push development

# Pull current schema from live DB into .sql files
skeema pull development
```

### How it works:
1. You define your tables in `.sql` files (CREATE TABLE statements)
2. Skeema compares them to the live database
3. It generates ALTER TABLE statements to bring the DB in sync

### vs EF Migrations (C#):
| Skeema | EF Core Migrations |
|--------|-------------------|
| Declarative (desired state) | Imperative (step-by-step changes) |
| `CREATE TABLE` = the truth | `migration_001_AddColumn.cs` |
| Diffs automatically | You write each migration |
| Like Terraform | Like CloudFormation |

### vs Flyway (Java):
| Skeema | Flyway |
|--------|--------|
| Schema diffing | Sequential migration files |
| One file per table | `V1__create_users.sql`, `V2__add_email.sql` |
| Order doesn't matter | Order is critical |
| Can "pull" from live DB | One-way only |

---

## Blank Imports (Driver Registration)

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"  // ← blank import
)
```

The `_` means "import for side effects only." The MySQL package's `init()` function registers the driver with `database/sql`. You never call it directly.

### C# Equivalent:
```csharp
// No equivalent needed — NuGet packages auto-register via DI
services.AddDbContext<AppContext>(o => o.UseMySql(...));
```

### Java Equivalent:
```java
// Modern JDBC uses ServiceLoader (automatic)
// Old way: Class.forName("com.mysql.cj.jdbc.Driver");
```

---

## The `defer` Pattern for Resource Cleanup

```go
rows, err := db.Query(...)
if err != nil { return err }
defer rows.Close()  // Executes when function returns (like finally)
```

### C# Equivalent:
```csharp
using var reader = await cmd.ExecuteReaderAsync();  // Disposed on scope exit
// or
await using var conn = new MySqlConnection(connStr);
```

### Java Equivalent:
```java
try (var rs = stmt.executeQuery()) {  // AutoCloseable
    while (rs.next()) { ... }
}
```

---

## Try It

The MySQL adapter compiles but requires a running MySQL instance. For local dev:

```bash
# Start MySQL via Docker
docker run -d --name shipit-mysql -p 3306:3306 \
    -e MYSQL_ROOT_PASSWORD=devpass \
    -e MYSQL_DATABASE=shipit \
    mysql:8.0

# Apply schema
# (install skeema from https://skeema.io)
cd schema
skeema push development

# Run with MySQL
$env:SHIPIT_DATABASE_URL = "root:devpass@tcp(127.0.0.1:3306)/shipit?parseTime=true"
.\shipit.exe serve
```

Without MySQL, the app runs fine with the in-memory adapter (it checks config to decide which adapter to use).

---

## Key Takeaways

1. **`database/sql`** — Go's stdlib for databases. Production-ready. Connection pooling built-in.
2. **Raw SQL** — no ORM. Write queries, scan results. Full control.
3. **`?` placeholders** — ALWAYS. Never concatenate user input. Prevents SQL injection.
4. **`defer rows.Close()`** — ALWAYS. Prevents connection pool exhaustion.
5. **`sql.ErrNoRows`** — sentinel error for "not found." Map it to your domain error.
6. **Blank import** (`_ "driver"`) — Go's plugin pattern for database drivers.
7. **Skeema** — declarative schema management (like Terraform for MySQL).

---

## Next: [Lesson 11 — Context & Cancellation](./11-context.md)
We'll add request-scoped context, timeouts, and graceful shutdown.
