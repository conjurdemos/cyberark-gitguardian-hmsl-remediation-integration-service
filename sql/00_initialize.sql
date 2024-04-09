-- 00_initialize.sql

CREATE DATABASE IF NOT EXISTS brimstone;

CREATE USER brimstone_dbuser WITH LOGIN PASSWORD 'Dev123';

GRANT ALL ON DATABASE brimstone TO brimstone_dbuser;

-- Recommended to change the password
-- https://www.cockroachlabs.com/docs/stable/alter-user#change-a-users-password