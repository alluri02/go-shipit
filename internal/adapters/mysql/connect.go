package mysql

import (
	"database/sql"
	"fmt"
	"time"

	// Blank import — registers the MySQL driver with database/sql.
	// This is Go's plugin pattern: import for side effects only.
	//
	// C# equivalent: No equivalent — EF/Dapper auto-discover providers.
	// Java equivalent: Class.forName("com.mysql.cj.jdbc.Driver") — old pattern.
	//                  Modern JDBC uses ServiceLoader (automatic).
	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a MySQL connection pool.
//
// KEY INSIGHT: sql.Open does NOT connect to the database.
// It only validates the DSN (Data Source Name) format.
// The actual connection is established lazily on first query.
// db.Ping() forces a connection check.
//
// C# equivalent:
//   var conn = new MySqlConnection(connectionString);
//   await conn.OpenAsync();  // connects immediately
//
// Java equivalent:
//   DataSource ds = HikariDataSourceBuilder.create()
//       .url(jdbcUrl).username(user).password(pass).build();
//   ds.getConnection();  // connects from pool
func Connect(dsn string) (*sql.DB, error) {
	// "mysql" is the driver name — registered by importing _ "github.com/go-sql-driver/mysql"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql.Connect: %w", err)
	}

	// Configure the connection pool
	db.SetMaxOpenConns(25)                 // Max simultaneous connections
	db.SetMaxIdleConns(5)                  // Keep 5 idle connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections every 5 min

	// Verify connectivity
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql.Connect ping: %w", err)
	}

	return db, nil
}
