-- Create a new schema
CREATE SCHEMA IF NOT EXISTS go_cache;
-- CREATE DATABASE go_cache;


-- Create table in specific schema
DROP TABLE IF EXISTS go_cache.workers;
CREATE TABLE go_cache.workers (
    id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    worker character varying(255) NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX idx_workers_created_at_desc ON go_cache.workers (created_at DESC);
CREATE INDEX idx_workers_updated_at_desc ON go_cache.workers (updated_at DESC);

-- List all schemas
-- SELECT schema_name FROM information_schema.schemata;

-- Set search path (schema precedence)
SET search_path TO go_cache, public;

-- List all tables in a schema
SELECT table_name FROM information_schema.tables WHERE table_schema = 'go_cache';