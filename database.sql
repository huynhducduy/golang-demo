create table `groups`
(
    id          int auto_increment,
    name        varchar(100) not null,
    description text         null,
    manager_id  int          null,
    constraint groups_id_uindex
        unique (id)
);

alter table `groups`
    add primary key (id);

create table users
(
    id        int auto_increment,
    username  varchar(100)         not null,
    password  varchar(255)         not null,
    group_id  int                  null,
    full_name varchar(255)         null,
    is_admin  tinyint(1) default 0 null,
    constraint users_id_uindex
        unique (id),
    constraint users_username_uindex
        unique (username),
    constraint users_groups_id_fk
        foreign key (group_id) references `groups` (id)
);

alter table users
    add primary key (id);

alter table `groups`
    add constraint groups_users_id_fk
        foreign key (manager_id) references users (id);

create table token
(
    `int`   int auto_increment,
    token   varchar(22) not null,
    user_id int         null,
    constraint token_int_uindex
        unique (`int`),
    constraint token_token_uindex
        unique (token),
    constraint token_users_id_fk
        foreign key (user_id) references users (id)
);

alter table token
    add primary key (`int`);

-- Cyclic dependencies found

alter table notifications
    add primary key (id);

-- Cyclic dependencies found

create table tasks
(
    id          int auto_increment,
    name        varchar(255)         not null,
    description text                 null,
    report      text                 null,
    assigner    int                  null,
    assignee    int                  null,
    review      int                  null,
    review_at   int                  null,
    comment     text                 null,
    proof       text                 null,
    start_at    int                  null,
    stop_at     int                  null,
    close_at    int                  null,
    open_at     int                  not null,
    open_from   int                  null,
    status      int        default 0 null,
    is_closed   tinyint(1) default 0 null,
    constraint tasks_id_uindex
        unique (id),
    constraint tasks_tasks_id_fk
        foreign key (open_from) references tasks (id),
    constraint tasks_users_id_fk
        foreign key (assignee) references users (id),
    constraint tasks_users_id_fk_2
        foreign key (assigner) references users (id)
);

alter table tasks
    add primary key (id);

create table notifications
(
    id      int auto_increment,
    user_id int                  not null,
    task_id int                  null,
    message text                 null,
    `read`  tinyint(1) default 0 null,
    constraint notifications_id_uindex
        unique (id),
    constraint notifications_tasks_id_fk
        foreign key (task_id) references tasks (id),
    constraint notifications_users_id_fk
        foreign key (user_id) references users (id)
);


