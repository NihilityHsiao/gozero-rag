CREATE TABLE `task` (
    `id` varchar(36) NOT NULL COMMENT '任务唯一ID (UUID)',
    `doc_id` varchar(36) NOT NULL COMMENT '关联的文档ID',
    `task_type` varchar(32) NOT NULL DEFAULT 'parse' COMMENT '任务类型: parse, graphrag,后期考虑实现raptor模型',
    `from_page` int(11) NOT NULL DEFAULT '0' COMMENT '起始页/行 (包含)',
    `to_page` int(11) NOT NULL DEFAULT '0' COMMENT '结束页/行 (不包含)',
    `progress` float NOT NULL DEFAULT '0' COMMENT '任务进度 0.0-1.0',
    `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态: pending | running | success | fail | paused',
    `progress_msg` text COMMENT '当前进度的详细日志/最后一条消息',
    `fail_reason` text COMMENT '如果失败，记录具体堆栈或错误信息',
    `retry_count` tinyint(3) NOT NULL DEFAULT '0' COMMENT '重试次数',-- 缓存与去重
    `digest` char(64) DEFAULT '' COMMENT '任务配置摘要Hash，用于检测重复任务/断点续传,如果一个文档的某个片段（分片），其内容范围和解析配置完全没有变，那么就直接复用上次的结果，不再重新跑 ORC/LLM。',
    `chunk_ids` longtext COMMENT '该任务生成的切片ID列表(空格或逗号分隔)，用于清理旧数据',
    `process_duration` float DEFAULT '0' COMMENT '处理耗时(秒)',
    `created_time` bigint NOT NULL COMMENT '创建时间戳(ms)',
    `updated_time` bigint NOT NULL COMMENT '更新时间戳(ms)',
    `created_date` datetime NOT NULL COMMENT '创建日期',
    `updated_date` datetime NOT NULL COMMENT '更新日期',
    `delete_at` timestamp NULL DEFAULT NULL COMMENT '软删除标记',
    PRIMARY KEY (`id`),
    KEY `idx_doc_id` (`doc_id`),
    KEY `idx_status` (`status`),
    KEY `idx_digest` (`digest`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档解析拆分任务表';