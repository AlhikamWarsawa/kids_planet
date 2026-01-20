INSERT INTO users (name, email, password_hash, role, status)
VALUES (
           'Admin',
           'admin@kidsplanet.com',
           '$2a$10$FuBgv/iPOnaSiQzbkEWWbOtlOz3bhsuYc5xVlkGFxVGEToro/qLQ2',
           'admin',
           'active'
       )
ON CONFLICT (email) DO NOTHING;