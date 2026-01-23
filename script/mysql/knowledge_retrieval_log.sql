use gozero_rag;

DROP TABLE IF EXISTS `knowledge_retrieval_log`;
CREATE TABLE `knowledge_retrieval_log`
(
    `id`               bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `knowledge_base_id` varchar(36) NOT NULL COMMENT '知识库ID (UUID)',
    `user_id`          varchar(36) NOT NULL COMMENT '用户ID (UUID)',
    `query`            text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户查询',
    `retrieval_mode`   varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '召回模式: vector, fulltext, hybrid',
    `retrieval_params` json DEFAULT NULL COMMENT '召回参数快照',
    `chunk_count`      int NOT NULL DEFAULT 0 COMMENT '召回片段数量',
    `time_cost_ms`     int NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
    `created_at`       datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_kb_id` (`knowledge_base_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库召回日志表';