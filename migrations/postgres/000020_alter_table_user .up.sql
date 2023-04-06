ALTER TABLE "user" DROP CONSTRAINT IF EXISTS user_unq_login;

ALTER TABLE "user" DROP CONSTRAINT IF EXISTS user_unq_phone;

<<<<<<< HEAD
CREATE UNIQUE INDEX IF NOT EXISTS user_unq_login ON "user" (login)
=======
CREATE UNIQUE INDEX if not EXISTS user_unq_login ON "user" (login)
>>>>>>> 89b750b653d9b0bdcee610905057880146f3dc29
WHERE(
   NOT (
     ( login IS NULL  OR  login = '' )
   )
);

<<<<<<< HEAD
CREATE UNIQUE INDEX IF NOT EXISTS user_unq_phone ON "user" (phone)
=======
CREATE UNIQUE INDEX if not EXISTS user_unq_phone ON "user" (phone)
>>>>>>> 89b750b653d9b0bdcee610905057880146f3dc29
WHERE(
   NOT (
     ( phone IS NULL  OR  phone = '' )
   )
);

<<<<<<< HEAD
CREATE UNIQUE INDEX IF NOT EXISTS user_login_and_phone_and_email ON "user" (phone, login, email)
=======
CREATE UNIQUE INDEX if not EXISTS user_login_and_phone_and_email ON "user" (phone, login, email)
>>>>>>> 89b750b653d9b0bdcee610905057880146f3dc29
WHERE (
   NOT (
     ( login IS NULL  OR  login = '' )
     AND
     ( phone IS NULL  OR  phone = '' )
   )
);