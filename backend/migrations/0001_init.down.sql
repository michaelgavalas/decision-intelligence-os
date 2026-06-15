-- Reverts migration 0001: drops identity and access primitives in reverse
-- dependency order and removes the citext extension.

DROP TABLE IF EXISTS rate_limits;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;
