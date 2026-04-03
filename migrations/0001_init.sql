-- +goose Up

CREATE TABLE IF NOT EXISTS links (
    id            bigserial primary key,
    original_url  text not null,
    short_code    varchar(10) not null,
    created_at    timestamptz not null default now(),

    constraint links_short_code_key unique (short_code),
    constraint links_original_url_key unique (original_url)
);

-- +goose Down

DROP TABLE IF EXISTS links;