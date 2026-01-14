use gozero_rag;
drop table if exists `tenant`;

-- 租户表
-- 一个租户可以看作一个team,那就可以用model_name作为model_id
CREATE TABLE `tenant` (
  `id` varchar(36) NOT NULL COMMENT '租户ID',
  `name` varchar(100) DEFAULT NULL COMMENT '租户名称',
--   `public_key` varchar(255) DEFAULT NULL COMMENT '公钥',
  `llm_id` varchar(128) NOT NULL default '' COMMENT '默认LLM模型ID',
  `embd_id` varchar(128) NOT NULL default '' COMMENT '默认Embedding模型ID',
  `asr_id` varchar(128) NOT NULL default '' COMMENT '默认ASR模型ID',
  `img2txt_id` varchar(128) NOT NULL default '' COMMENT '默认图片转文字模型ID',
  `rerank_id` varchar(128) NOT NULL default '' COMMENT '默认重排序模型ID',
  `tts_id` varchar(256) NOT NULL default '' COMMENT '默认TTS模型ID',
  `parser_ids` varchar(256) NOT NULL default '' COMMENT '文档处理器列表',
  `credit` int NOT NULL DEFAULT 512 COMMENT '积分',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1=正常, 0=禁用',
  `created_time` bigint NOT NULL COMMENT '创建时间戳',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';