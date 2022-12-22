create table if not exists login_strategy (
    id uuid primary key,
    name varchar(255),
    view_field varchar(255)
);

alter table "user"
    drop column if exists client_platform_id;

alter table "user"
    drop column if exists client_type_id;

alter table "user"
    drop column if exists role_id;

alter table "user"
    drop column if exists project_id;

alter table "user"
    add column if not exists project_id uuid;