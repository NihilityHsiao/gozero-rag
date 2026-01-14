use gozero_rag;
drop table if exists `tenant_llm`;

-- 租户LLM配置表 (复合主键)
CREATE TABLE `tenant_llm` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `tenant_id` varchar(36) NOT NULL COMMENT '租户ID',
  `llm_factory` varchar(128) NOT NULL COMMENT 'LLM厂商名称',
  `model_type` varchar(128) DEFAULT NULL COMMENT '模型类型: LLM, Text Embedding, Image2Text, ASR',
  `llm_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'LLM模型名称',
  `api_key` longtext COMMENT 'API密钥',
  `api_base` varchar(255) DEFAULT NULL COMMENT 'API基础地址',
  `max_tokens` int NOT NULL DEFAULT 8192 COMMENT '最大Token数',
  `used_tokens` int NOT NULL DEFAULT 0 COMMENT '已使用Token数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0-废弃, 1-有效',
  `created_time` bigint NOT NULL COMMENT '创建时间戳',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_factory_llm` (`tenant_id`, `llm_factory`, `llm_name`),
  KEY `idx_tenant_id` (`tenant_id`),
  KEY `idx_llm_factory` (`llm_factory`),
  KEY `idx_model_type` (`model_type`),
  KEY `idx_llm_name` (`llm_name`),
  KEY `idx_max_tokens` (`max_tokens`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户LLM配置表';