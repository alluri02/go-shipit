-- Schema for the deployments table.
-- Managed by Skeema (schema diffing tool) — NOT migration files.
--
-- Skeema compares this file to the live database and generates ALTER statements.
-- Think of it as "Terraform for MySQL schemas."
--
-- C# equivalent: EF Core Migrations (Add-Migration, Update-Database)
-- Java equivalent: Flyway / Liquibase migration files

CREATE TABLE deployments (
    id VARCHAR(64) NOT NULL,
    service_name VARCHAR(128) NOT NULL,
    image_tag VARCHAR(128) NOT NULL,
    environment VARCHAR(32) NOT NULL,
    region VARCHAR(32) NOT NULL,
    cluster_url VARCHAR(512) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 0,
    risk_score TINYINT NOT NULL DEFAULT 0,
    triggered_by VARCHAR(32) NOT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_service_created (service_name, created_at DESC),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
