DROP TABLE IF EXISTS tour_versions;

DROP TRIGGER IF EXISTS tours_set_updated_at ON tours;
DROP TABLE IF EXISTS tours;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS apps;