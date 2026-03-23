DROP INDEX IF EXISTS bookings_created_at_idx;
DROP INDEX IF EXISTS bookings_user_created_at_idx;
DROP INDEX IF EXISTS slots_room_start_idx;
DROP INDEX IF EXISTS bookings_active_slot_uidx;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
