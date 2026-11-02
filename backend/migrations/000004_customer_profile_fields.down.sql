ALTER TABLE users
    DROP COLUMN IF EXISTS platform,
    DROP COLUMN IF EXISTS os_version,
    DROP COLUMN IF EXISTS device_model,
    DROP COLUMN IF EXISTS device_brand,
    DROP COLUMN IF EXISTS country,
    DROP COLUMN IF EXISTS birthdate,
    DROP COLUMN IF EXISTS gender;
