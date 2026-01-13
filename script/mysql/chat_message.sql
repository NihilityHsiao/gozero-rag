use gozero_rag;


-- ----------------------------
-- Table structure for chat_message
-- ----------------------------
DROP TABLE IF EXISTS `chat_message`;
CREATE TABLE `chat_message` (
  `id` char(36) NOT NULL COMMENT '消息ID (UUID)',
  `conversation_id` char(36) NOT NULL COMMENT '所属会话ID',
  `seq_id` int(11) NOT NULL DEFAULT '0' COMMENT '消息序号 (用于严格排序)',
  `role` varchar(32) NOT NULL COMMENT '角色: user, assistant, system, tool',
  `content` longtext NOT NULL COMMENT '消息内容',
  `type` varchar(32) NOT NULL DEFAULT 'text' COMMENT '内容类型: text, json (for tool calls)',
  `token_count` int(11) NOT NULL DEFAULT '0' COMMENT 'Token数量估算',
  `expert_mode` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否专家模式/深度思考模式',
  `model_config` json DEFAULT NULL COMMENT '当前消息使用的模型配置(Snapshot): model_id, knowledge_base_ids, etc.',
  `extra` json DEFAULT NULL COMMENT '扩展信息: 引用源(citations), 思考过程(reasoning), 耗时等',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_conversation_seq` (`conversation_id`,`seq_id`),
  KEY `idx_project_msg` (`conversation_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';
