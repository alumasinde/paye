-- Seeds the first admin account so the admin panel can be logged into
-- immediately after migrating. Change this password after first login -
-- there is currently no "change password" endpoint, only direct DB access
-- (UPDATE admin_users SET password_hash=... WHERE email=...) with a bcrypt
-- hash generated the same way this one was.

INSERT INTO admin_users (public_id, email, password_hash, first_name, last_name, status)
SELECT UUID(), 'alumasinde@gmail.com', '$2a$10$MiOa4/3UbMZ11GrwrvKvDub.cD/FLByLfPYoq9danShdWnO8Jwkoq', 'Albert', 'Masinde', 'ACTIVE'
WHERE NOT EXISTS (SELECT 1 FROM admin_users WHERE email = 'alumasinde@gmail.com');

INSERT INTO admin_user_roles (admin_user_id, role_id)
SELECT au.id, ar.id
FROM admin_users au
JOIN admin_roles ar ON ar.code = 'SUPER_ADMIN'
WHERE au.email = 'alumasinde@gmail.com'
  AND NOT EXISTS (
    SELECT 1 FROM admin_user_roles x WHERE x.admin_user_id = au.id AND x.role_id = ar.id
  );
