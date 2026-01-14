use gozero_rag;
drop table if exists `user_tenant`;

-- 用户租户关联表
CREATE TABLE `user_tenant` (
  `id` varchar(36) NOT NULL COMMENT '记录ID',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `role` varchar(32) NOT NULL COMMENT '角色: UserTenantRole',
  `invited_by` varchar(36) NOT NULL COMMENT '邀请人ID',
  `status` tinyint DEFAULT 1 COMMENT '状态: 0-废弃, 1-有效',
  `created_time` bigint NOT NULL COMMENT '创建时间戳(ms)',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳(ms)',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户租户关联表';