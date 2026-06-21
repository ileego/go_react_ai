DROP TABLE IF EXISTS files;
DROP INDEX IF EXISTS idx_reports_search;
DROP TRIGGER IF EXISTS reports_search_trigger ON reports;
DROP FUNCTION IF EXISTS reports_search_update();
ALTER TABLE reports DROP COLUMN IF EXISTS search_vector;
