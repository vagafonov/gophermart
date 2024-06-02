create table users
(
    id     uuid not null,
    login  VARCHAR(255),
    password VARCHAR(60)
);

create unique index users_login_uindex on users (login);


