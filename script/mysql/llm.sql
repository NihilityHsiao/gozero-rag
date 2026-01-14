use gozero_rag;
drop table if exists `llm`;

-- LLM 模型字典表 (复合主键)
CREATE TABLE `llm` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `llm_name` varchar(128) NOT NULL COMMENT 'LLM模型名称',
  `model_type` varchar(128) NOT NULL COMMENT '模型类型: LLM, Text Embedding, Image2Text, ASR',
  `fid` varchar(128) NOT NULL COMMENT 'LLM厂商ID',
  `max_tokens` int NOT NULL DEFAULT 0 COMMENT '最大Token数',
  `tags` varchar(255) NOT NULL COMMENT '标签: LLM, Text Embedding, Image2Text, Chat, 32k...',
  `is_tools` tinyint NOT NULL DEFAULT 0 COMMENT '是否支持工具调用',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0-废弃, 1-有效',
  `created_time` bigint NOT NULL COMMENT '创建时间戳',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_fid_llm_name` (`fid`, `llm_name`),
  KEY `idx_llm_name` (`llm_name`),
  KEY `idx_model_type` (`model_type`),
  KEY `idx_fid` (`fid`),
  KEY `idx_tags` (`tags`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='LLM模型字典表';