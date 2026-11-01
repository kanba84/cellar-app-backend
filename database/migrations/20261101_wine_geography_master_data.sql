BEGIN;

ALTER TABLE regions
    ADD COLUMN IF NOT EXISTS code varchar(64),
    ADD COLUMN IF NOT EXISTS sort_order integer NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_regions_code ON regions (code);

ALTER TABLE appellations
    ADD COLUMN IF NOT EXISTS code varchar(100),
    ADD COLUMN IF NOT EXISTS country_id integer,
    ADD COLUMN IF NOT EXISTS sort_order integer NOT NULL DEFAULT 0;

UPDATE appellations AS a
SET country_id = r.country_id
FROM regions AS r
WHERE a.region_id = r.id
  AND a.country_id IS NULL;

ALTER TABLE appellations
    DROP CONSTRAINT IF EXISTS appellations_country_id_fkey;

ALTER TABLE appellations
    ADD CONSTRAINT appellations_country_id_fkey
    FOREIGN KEY (country_id) REFERENCES countries (id) ON DELETE NO ACTION;

CREATE INDEX IF NOT EXISTS idx_appellations_country_id ON appellations (country_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_appellations_code ON appellations (code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_countries_iso_code ON countries (iso_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_designation_types_country_code
    ON designation_types (code, country_id);

CREATE TABLE IF NOT EXISTS appellation_regions (
    appellation_id integer NOT NULL,
    region_id integer NOT NULL,
    PRIMARY KEY (appellation_id, region_id),
    CONSTRAINT fk_appellation_regions_appellation
        FOREIGN KEY (appellation_id) REFERENCES appellations (id) ON DELETE CASCADE,
    CONSTRAINT fk_appellation_regions_region
        FOREIGN KEY (region_id) REFERENCES regions (id) ON DELETE CASCADE
);

ALTER TABLE appellation_regions
    DROP CONSTRAINT IF EXISTS fk_appellation_regions_appellation;

ALTER TABLE appellation_regions
    ADD CONSTRAINT fk_appellation_regions_appellation
    FOREIGN KEY (appellation_id) REFERENCES appellations (id) ON DELETE CASCADE;

ALTER TABLE appellation_regions
    DROP CONSTRAINT IF EXISTS fk_appellation_regions_region;

ALTER TABLE appellation_regions
    ADD CONSTRAINT fk_appellation_regions_region
    FOREIGN KEY (region_id) REFERENCES regions (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_appellation_regions_region_id
    ON appellation_regions (region_id);

INSERT INTO appellation_regions (appellation_id, region_id)
SELECT id, region_id
FROM appellations
WHERE region_id IS NOT NULL
ON CONFLICT DO NOTHING;

ALTER TABLE appellations
    DROP CONSTRAINT IF EXISTS appellations_region_id_fkey;

ALTER TABLE appellations
    ADD CONSTRAINT appellations_region_id_fkey
    FOREIGN KEY (region_id) REFERENCES regions (id) ON DELETE RESTRICT;

COMMIT;
