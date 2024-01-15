-- 001_create_users_table.up.sql
-- Create the tigerhall schema
CREATE SCHEMA IF NOT EXISTS tigerhall;
-- Create the users table
CREATE TABLE IF NOT EXISTS tigerhall.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL
);


-- 002_create_tigers_table.up.sql
-- Create the tigers table
CREATE TABLE IF NOT EXISTS tigerhall.tigers (
    tiger_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    date_of_birth DATE NOT NULL,
    last_seen_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_coordinates_lat DOUBLE PRECISION NOT NULL,
    last_seen_coordinates_lon DOUBLE PRECISION NOT NULL
);

-- Create an index on the last_seen_timestamp column in tigers table for sorting
CREATE INDEX IF NOT EXISTS idx_tigers_last_seen_timestamp ON tigerhall.tigers(last_seen_timestamp);


-- 003_create_sightings_table.up.sql
-- Create the sightings table within the tigerhall schema
CREATE TABLE IF NOT EXISTS tigerhall.sightings (
    sighting_id SERIAL PRIMARY KEY,
    tiger_id INT REFERENCES tigerhall.tigers(tiger_id) ON DELETE CASCADE,
    user_id INT REFERENCES tigerhall.users(user_id) ON DELETE CASCADE,
    last_seen_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen_coordinates_lat DOUBLE PRECISION NOT NULL,
    last_seen_coordinates_lon DOUBLE PRECISION NOT NULL,
    image BYTEA
);

-- Create an index on the timestamp column in sightings table for sorting
CREATE INDEX IF NOT EXISTS idx_sightings_timestamp ON tigerhall.sightings(last_seen_timestamp);
