-- Renames the remaining Spanish identifiers to English:
--   platform.anden -> platform.bay
--   rol            -> role
--   users.rol_id   -> users.role_id
-- and renames the matching keys inside stored notification payloads.

ALTER TABLE platform RENAME COLUMN anden TO bay;

ALTER TABLE rol RENAME TO role;
ALTER TABLE users RENAME COLUMN rol_id TO role_id;
ALTER INDEX IF EXISTS idx_users_rol_id RENAME TO idx_users_role_id;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rol_pkey') THEN
        ALTER TABLE role RENAME CONSTRAINT rol_pkey TO role_pkey;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rol_name_key') THEN
        ALTER TABLE role RENAME CONSTRAINT rol_name_key TO role_name_key;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_rol_id_fkey') THEN
        ALTER TABLE users RENAME CONSTRAINT users_rol_id_fkey TO users_role_id_fkey;
    END IF;
    -- PostgreSQL 18+ also stores NOT NULL constraints by name; older versions have none.
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rol_uuid_not_null') THEN
        ALTER TABLE role RENAME CONSTRAINT rol_uuid_not_null TO role_uuid_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rol_name_not_null') THEN
        ALTER TABLE role RENAME CONSTRAINT rol_name_not_null TO role_name_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_rol_id_not_null') THEN
        ALTER TABLE users RENAME CONSTRAINT users_rol_id_not_null TO users_role_id_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'platform_anden_not_null') THEN
        ALTER TABLE platform RENAME CONSTRAINT platform_anden_not_null TO platform_bay_not_null;
    END IF;
END $$;

UPDATE notifications
SET payload = jsonb_set(payload #- '{payload,anden}', '{payload,bay}', payload -> 'payload' -> 'anden')
WHERE payload -> 'payload' ? 'anden';

UPDATE notifications
SET payload = jsonb_set(payload #- '{payload,license_patent}', '{payload,license_plate}', payload -> 'payload' -> 'license_patent')
WHERE payload -> 'payload' ? 'license_patent';
