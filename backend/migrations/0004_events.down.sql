-- Reverts migration 0004: drops the audit event log table.

DROP TABLE IF EXISTS events;
