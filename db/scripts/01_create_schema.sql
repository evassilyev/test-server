create table if not exists balance_history (
    date_time timestamp not null default now(),
    operation char(4) not null,
    amount numeric not null,
    tid varchar not null unique,
    deleted bool not null default false
);
create unique index if not exists balance_tid_idx on balance_history (tid);

drop view if exists calculated_balance_view;
create view calculated_balance_view(balance) as
    select coalesce(sum(amnt), 0) as balance from
        (select
                (case operation
                    when 'win' then amount
                    when 'lost' then -amount
                end) as amnt
        from balance_history where deleted = false) as amounts;

