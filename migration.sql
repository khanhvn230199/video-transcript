-- tạo bảng users
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,

    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,

    name VARCHAR(255),
    avatar_url TEXT,

    role VARCHAR(20) NOT NULL DEFAULT 'user',

    credit INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- tạo bảng videos
CREATE TABLE videos (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL,          -- ID người dùng
    link_video TEXT NOT NULL,         -- URL video (R2 URL hoặc custom domain)
    name_file TEXT NOT NULL,          -- tên file gốc (vd: myvideo.mp4)
    description TEXT,                 -- mô tả video

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);