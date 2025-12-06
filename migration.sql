-- tạo bảng users
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,

    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,

    name VARCHAR(255),
    avatar_url TEXT,

    gender VARCHAR(10),
    dob DATE,
    phone VARCHAR(20),
    address TEXT,

    role VARCHAR(20) NOT NULL DEFAULT 'user',

    credit INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- tạo bảng videos
CREATE TABLE IF NOT EXISTS videos (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL,          -- ID người dùng
    link_video TEXT NOT NULL,         -- URL video (R2 URL hoặc custom domain)
    name_file TEXT NOT NULL,          -- tên file gốc (vd: myvideo.mp4)
    description TEXT,                 -- mô tả video

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE tasks (
    id              BIGSERIAL PRIMARY KEY,

    task_type       VARCHAR(10),
    status_task          VARCHAR(10) NOT NULL DEFAULT 'pending',

    input_text      TEXT,
    input_url       TEXT,
    output_url      TEXT,

    transcript_text TEXT,
    transcript_json JSONB,

    duration_sec    FLOAT,
    error_message   TEXT,
    user_id         BIGINT,

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================
-- MIGRATION: Thêm các trường mới vào bảng users
-- Chạy các lệnh ALTER TABLE bên dưới nếu database đã có dữ liệu
-- ============================================

-- Thêm cột gender (giới tính) - chỉ chạy nếu cột chưa tồn tại
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS gender VARCHAR(10);

-- Thêm cột dob (ngày sinh) - chỉ chạy nếu cột chưa tồn tại
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS dob DATE;

-- Thêm cột phone (số điện thoại) - chỉ chạy nếu cột chưa tồn tại
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- Thêm cột address (địa chỉ) - chỉ chạy nếu cột chưa tồn tại
-- ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT;

-- Hoặc chạy tất cả cùng lúc (PostgreSQL 9.5+):
-- ALTER TABLE users 
--     ADD COLUMN IF NOT EXISTS gender VARCHAR(10),
--     ADD COLUMN IF NOT EXISTS dob DATE,
--     ADD COLUMN IF NOT EXISTS phone VARCHAR(20),
--     ADD COLUMN IF NOT EXISTS address TEXT;


