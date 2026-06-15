-- Reverts migration 0002: drops the decision core tables in reverse
-- dependency order.

DROP TABLE IF EXISTS evidence;
DROP TABLE IF EXISTS assumptions;
DROP TABLE IF EXISTS decisions;
