alter table "user"
add constraint user_unq unique(login);

alter table "user"
add constraint user_unq unique(email);

alter table "user"
add constraint user_unq unique(phone);