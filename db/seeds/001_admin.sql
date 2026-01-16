INSERT INTO users (name, email, password_hash, role, status)
VALUES (
           'Admin',
           'admin@kidsplanet.com',
           '123456',
           'admin',
           'active'
       )
ON CONFLICT (email) DO NOTHING;
