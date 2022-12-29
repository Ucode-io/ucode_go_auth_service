ALTER TABLE auth_service.user DROP COLUMN  IF EXISTS project_id;

ALTER TABLE auth_service.user ADD COLUMN company_id uuid;