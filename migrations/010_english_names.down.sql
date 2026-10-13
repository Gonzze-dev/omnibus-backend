UPDATE notifications
SET payload = jsonb_set(payload #- '{payload,license_plate}', '{payload,license_patent}', payload -> 'payload' -> 'license_plate')
WHERE payload -> 'payload' ? 'license_plate';

UPDATE notifications
SET payload = jsonb_set(payload #- '{payload,bay}', '{payload,anden}', payload -> 'payload' -> 'bay')
WHERE payload -> 'payload' ? 'bay';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'platform_bay_not_null') THEN
        ALTER TABLE platform RENAME CONSTRAINT platform_bay_not_null TO platform_anden_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_id_not_null') THEN
        ALTER TABLE users RENAME CONSTRAINT users_role_id_not_null TO users_rol_id_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'role_name_not_null') THEN
        ALTER TABLE role RENAME CONSTRAINT role_name_not_null TO rol_name_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'role_uuid_not_null') THEN
        ALTER TABLE role RENAME CONSTRAINT role_uuid_not_null TO rol_uuid_not_null;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_id_fkey') THEN
        ALTER TABLE users RENAME CONSTRAINT users_role_id_fkey TO users_rol_id_fkey;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'role_name_key') THEN
        ALTER TABLE role RENAME CONSTRAINT role_name_key TO rol_name_key;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'role_pkey') THEN
        ALTER TABLE role RENAME CONSTRAINT role_pkey TO rol_pkey;
    END IF;
END $$;

ALTER INDEX IF EXISTS idx_users_role_id RENAME TO idx_users_rol_id;
ALTER TABLE users RENAME COLUMN role_id TO rol_id;
ALTER TABLE role RENAME TO rol;

ALTER TABLE platform RENAME COLUMN bay TO anden;
