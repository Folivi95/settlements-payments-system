-- create database
CREATE DATABASE settlements_payments;

-- create readonly role
CREATE ROLE readonly;
GRANT CONNECT ON DATABASE settlements_payments TO readonly;
GRANT USAGE ON SCHEMA public TO readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly;


-- create read-write role
CREATE ROLE readwrite;
GRANT CONNECT ON DATABASE settlements_payments TO readwrite;
GRANT USAGE, CREATE ON SCHEMA public TO readwrite;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO readwrite;


-- create users
CREATE USER settlements_payments_system WITH PASSWORD 'XXXXXXXXXXXX';
GRANT readwrite TO settlements_payments_system;