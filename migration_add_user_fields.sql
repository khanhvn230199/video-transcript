-- Migration: Thêm các trường mới vào bảng users
-- Chạy file này nếu database đã có dữ liệu và cần thêm các cột: gender, dob, phone, address
-- Date: 2024

-- Thêm cột gender (giới tính)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'gender'
    ) THEN
        ALTER TABLE users ADD COLUMN gender VARCHAR(10);
        RAISE NOTICE 'Added column: gender';
    ELSE
        RAISE NOTICE 'Column gender already exists';
    END IF;
END $$;

-- Thêm cột dob (ngày sinh)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'dob'
    ) THEN
        ALTER TABLE users ADD COLUMN dob DATE;
        RAISE NOTICE 'Added column: dob';
    ELSE
        RAISE NOTICE 'Column dob already exists';
    END IF;
END $$;

-- Thêm cột phone (số điện thoại)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'phone'
    ) THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(20);
        RAISE NOTICE 'Added column: phone';
    ELSE
        RAISE NOTICE 'Column phone already exists';
    END IF;
END $$;

-- Thêm cột address (địa chỉ)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'address'
    ) THEN
        ALTER TABLE users ADD COLUMN address TEXT;
        RAISE NOTICE 'Added column: address';
    ELSE
        RAISE NOTICE 'Column address already exists';
    END IF;
END $$;

-- Xác nhận migration hoàn tất
SELECT 'Migration completed: Added gender, dob, phone, address columns to users table' AS status;

