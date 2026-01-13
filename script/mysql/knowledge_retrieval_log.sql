use gozero_rag;
drop table if exists `knowledge_retrieval_log`;

CREATE TABLE `knowledge_retrieval_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(11) not null  COMMENT '用户唯一标识',
  `knowledge_base_id` bigint(20) unsigned NOT NULL COMMENT '知识库ID',
  `query` varchar(2048) NOT NULL DEFAULT '' COMMENT '用户原始查询',
  `retrieval_mode` varchar(2048) DEFAULT '' COMMENT '检索模式 vector/fulltext/hybrid',
  `retrieval_params` json DEFAULT NULL COMMENT '检索参数快照(阈值/权重等)',
  `chunk_count` int(11) NOT NULL DEFAULT '0' COMMENT '召回chunk数',
  `time_cost_ms` int(11) NOT NULL DEFAULT '0' COMMENT '查询耗时(ms)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
  INDEX `idx_kb_created` (`knowledge_base_id`, `created_at`)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='RAG-知识库召回日志表';