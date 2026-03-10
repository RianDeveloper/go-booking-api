create extension if not exists "pgcrypto";

create table if not exists users (
    id uuid primary key,
    email text not null unique,
    name text not null,
    created_at timestamptz not null
);

create table if not exists resources (
    id uuid primary key,
    name text not null,
    created_at timestamptz not null
);

create table if not exists slots (
    id uuid primary key,
    resource_id uuid not null references resources(id) on delete cascade,
    start_time timestamptz not null,
    end_time timestamptz not null,
    status text not null check (status in ('free', 'booked')),
    version bigint not null,
    created_at timestamptz not null,
    unique(resource_id, start_time, end_time)
);

create table if not exists bookings (
    id uuid primary key,
    user_id uuid not null references users(id) on delete restrict,
    resource_id uuid not null references resources(id) on delete restrict,
    slot_id uuid not null references slots(id) on delete restrict,
    status text not null check (status in ('created', 'cancelled')),
    created_at timestamptz not null,
    cancelled_at timestamptz null
);

create table if not exists outbox_events (
    id uuid primary key,
    event_type text not null,
    aggregate_id uuid not null,
    payload jsonb not null,
    status text not null check (status in ('pending', 'published')),
    attempt integer not null default 0,
    next_attempt_at timestamptz not null default now(),
    created_at timestamptz not null,
    published_at timestamptz null
);

create index if not exists idx_slots_resource_time on slots(resource_id, start_time, end_time);
create index if not exists idx_outbox_pending on outbox_events(status, next_attempt_at, created_at);
