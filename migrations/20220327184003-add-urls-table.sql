-- +migrate Up
create table if not exists urls
(
    id              uuid default gen_random_uuid(),
    short           text not null,
    original        text not null,
    user_id         text not null,
    created_at      timestamp default now(),

    constraint urls_pk primary key (id),
    constraint original unique (original)
);
-- +migrate Down
drop table urls;
