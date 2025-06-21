-- Auth Service Database Migration Script
CREATE DATABASE auth_db;
CREATE USER auth_service WITH PASSWORD 'my-auth-secret-password';
GRANT CONNECT ON DATABASE auth_db TO auth_service;
GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth_service;
GRANT USAGE ON SCHEMA public TO auth_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO auth_service;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO auth_service;

-- Library Service Database Migration Script
CREATE DATABASE library_db;
CREATE USER library_service WITH PASSWORD 'my-library-secret-password';
GRANT CONNECT ON DATABASE library_db TO library_service;
GRANT ALL PRIVILEGES ON DATABASE library_db TO library_service;
GRANT USAGE ON SCHEMA public TO library_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO library_service;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO library_service;

-- Transaction Service Database Migration Script
CREATE DATABASE transaction_db;
CREATE USER transaction_service WITH PASSWORD 'my-transaction-secret-password';
GRANT CONNECT ON DATABASE transaction_db TO transaction_service;
GRANT ALL PRIVILEGES ON DATABASE transaction_db TO transaction_service;
GRANT USAGE ON SCHEMA public TO transaction_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO transaction_service;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO transaction_service;