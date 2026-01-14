use `gozero_rag`;
drop table if exists `user`;

-- 使用jwt + redis实现用户认证
CREATE TABLE `user` (
  `id` varchar(36) NOT NULL COMMENT '用户ID, UUID',
  `nickname` varchar(100) NOT NULL COMMENT '昵称',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `email` varchar(255) NOT NULL COMMENT '邮箱',
  `avatar` longtext COMMENT '头像base64字符串',
  `language` varchar(32) DEFAULT 'Chinese' COMMENT '语言: English|Chinese',
  `color_schema` varchar(32) DEFAULT 'Bright' COMMENT '主题: Bright|Dark',
  `timezone` varchar(64) DEFAULT 'UTC+8\tAsia/Shanghai' COMMENT '时区',
  `last_login_time` datetime DEFAULT NULL COMMENT '最后登录时间',
  `is_active` tinyint NOT NULL DEFAULT 1 COMMENT '是否激活,核心业务开关,只允许值为1的用户登录',
  `login_channel` varchar(255) DEFAULT NULL COMMENT '登录渠道',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0-废弃, 1-有效',
  `is_superuser` tinyint NOT NULL DEFAULT 0 COMMENT '是否超级管理员',
  `created_time` bigint NOT NULL COMMENT '创建时间戳(ms)',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳(ms)',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`),
  KEY `idx_nickname` (`nickname`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`created_time`),
  KEY `idx_update_time` (`updated_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
