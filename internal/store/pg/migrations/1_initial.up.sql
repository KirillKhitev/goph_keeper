CREATE TABLE users(id varchar(36) PRIMARY KEY, user_name varchar(255), hash_password varchar(255), deleted boolean DEFAULT FALSE, registration_date timestamp);
CREATE TABLE IF NOT EXISTS datas (id varchar(36) PRIMARY KEY, name bytea DEFAULT NULL, user_id varchar(36) NOT NULL, type varchar(100) NOT NULL, date timestamp, body bytea DEFAULT NULL, deleted boolean DEFAULT FALSE, description bytea DEFAULT NULL);
CREATE UNIQUE INDEX IF NOT EXISTS user_name_users_idx ON users (user_name);
CREATE INDEX IF NOT EXISTS del_user_type_idx ON datas (deleted, user_id, type);