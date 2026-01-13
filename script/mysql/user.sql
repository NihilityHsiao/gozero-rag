use `gozero_rag`;

create table `user` (
  `id` int(11) not null auto_increment COMMENT '用户唯一标识，自增主键',
  `username` varchar(255) not null COMMENT '用户名，用于登录，唯一',
  `password_hash` varchar(255) not null COMMENT '密码哈希值，存储加密后的密码',
  `email` varchar(255) not null COMMENT '用户邮箱地址',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    primary key (`id`),
    unique key `uk_username` (`username`) COMMENT '用户名唯一索引，确保用户名不重复',
    KEY `idx_created_at` (`created_at`) USING BTREE COMMENT '创建时间索引（按注册时间筛选）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户信息表';