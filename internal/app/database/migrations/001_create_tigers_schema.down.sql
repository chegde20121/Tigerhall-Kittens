-- 002__create_users_table.down.sql
-- Drop the users table
-- Drop the sightings table
DROP TABLE IF EXISTS tigerhall.sightings;

DROP TABLE IF EXISTS tigerhall.tigers;

DROP TABLE IF EXISTS tigerhall.users;

DROP SCHEMA IF EXISTS tigerhall;