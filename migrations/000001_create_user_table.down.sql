-- migrations/000001_create_users_table.down.sql
DROP INDEX IF EXISTS idx_login;
DROP TABLE IF EXISTS users;