CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    passwordHash TEXT NOT NULL,
    firstName TEXT NOT NULL,
    lastName TEXT NOT NULL,
    phone TEXT.
    role TEXT NOT NULL CHECK (role IN ('customer', 'admin')) DEFAULT 'customer'
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);