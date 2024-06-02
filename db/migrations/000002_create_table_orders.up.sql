create table orders
(
    id         varchar      not null constraint orders_pk primary key,
    user_id    uuid      not null,
    status     smallint  not null,
    type smallint,
    amount decimal,
    created_at timestamp not null,
    updated_at timestamp
);

create index orders_status_index
    on orders (status);

