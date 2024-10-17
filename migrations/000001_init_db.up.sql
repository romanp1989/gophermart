create table if not exists users
(
    id       bigserial
        constraint users_pk
            primary key,
    login    varchar(255) not null,
    password varchar(255) not null
);

create unique index if not exists users_login_udx
    on users (login, login);

create table if not exists orders
(
    id         bigserial
        constraint orders_pk
            primary key,
    created_at timestamp with time zone default now() not null,
    number     varchar(100)                           not null,
    status     varchar(20) not null,
    user_id    bigint
        constraint orders_user_id_idx
            references users
);

create unique index if not exists orders_number_udx
    on orders (number);

create table if not exists balance
(
    id           bigserial
        constraint balance_pk
            primary key,
    created_at   timestamp with time zone not null,
    order_number varchar(50)              not null,
    user_id      bigint                   not null
        constraint balance_user_id_fk
            references public.users,
    sum          numeric(15, 2)           not null,
    type         smallint                 not null
);

create index if not exists balance_user_id_idx
    on balance (user_id);
