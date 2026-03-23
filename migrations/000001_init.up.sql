CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT users_email_not_blank CHECK (char_length(btrim(email)) > 0),
    CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'))
);

INSERT INTO users (id, email, password_hash, role, created_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'dummy-admin@example.com', NULL, 'admin', NOW()),
    ('00000000-0000-0000-0000-000000000002', 'dummy-user@example.com', NULL, 'user', NOW());

CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    capacity INTEGER,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT rooms_name_not_blank CHECK (char_length(btrim(name)) > 0),
    CONSTRAINT rooms_capacity_positive CHECK (capacity IS NULL OR capacity > 0)
);

CREATE TABLE schedules (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week SMALLINT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    CONSTRAINT schedules_days_not_empty CHECK (cardinality(days_of_week) > 0),
    CONSTRAINT schedules_days_range CHECK (days_of_week <@ ARRAY[1,2,3,4,5,6,7]::SMALLINT[]),
    CONSTRAINT schedules_time_order CHECK (start_time < end_time),
    CONSTRAINT schedules_time_range_min_slot CHECK (EXTRACT(EPOCH FROM (end_time - start_time)) >= 1800),
    CONSTRAINT schedules_time_range_aligned CHECK (MOD(EXTRACT(EPOCH FROM (end_time - start_time))::INTEGER, 1800) = 0)
);

CREATE TABLE slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT slots_time_order CHECK (start_at < end_at),
    CONSTRAINT slots_fixed_duration CHECK (EXTRACT(EPOCH FROM (end_at - start_at)) = 1800),
    CONSTRAINT slots_room_start_unique UNIQUE (room_id, start_at)
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL,
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT bookings_status_check CHECK (status IN ('active', 'cancelled'))
);

CREATE UNIQUE INDEX bookings_active_slot_uidx
    ON bookings (slot_id)
    WHERE status = 'active';

CREATE INDEX slots_room_start_idx
    ON slots (room_id, start_at);

CREATE INDEX bookings_user_created_at_idx
    ON bookings (user_id, created_at DESC);

CREATE INDEX bookings_created_at_idx
    ON bookings (created_at DESC);
