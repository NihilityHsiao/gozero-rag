use gozero_rag;

-- ----------------------------
-- Table structure for chat_conversation
-- ----------------------------
DROP TABLE IF EXISTS `chat_conversation`;
CREATE TABLE `chat_conversation` (
  `id` char(36) NOT NULL COMMENT '会话ID (UUID)',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID (UUID)',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID (UUID)',
  `title` varchar(255) NOT NULL DEFAULT 'New Conversation' COMMENT '会话标题',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 1-正常, 2-归档, 3-删除',
  `config` json DEFAULT NULL COMMENT '对话配置: llm_id, system_prompt, etc.',
  `message_count` int(11) NOT NULL DEFAULT '0' COMMENT '消息数量',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_user_updated` (`user_id`,`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';

