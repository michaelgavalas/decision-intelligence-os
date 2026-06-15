-- Reverts migration 0003: drops forecasting and resolution tables in reverse
-- dependency order.

DROP TABLE IF EXISTS outcomes;
DROP TABLE IF EXISTS predictions;
