create table `local_message` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',

    `task_type` varchar(64) not null COMMENT '任务类型, document_index | graph_extraction',
    `req_snapshot` longtext NOT NULL COMMENT '请求快照 json',
    `status` varchar(16) not null COMMENT '状态, init | fail | success',
    `next_retry_time` bigint not null COMMENT '下一次重试时间',
    `retry_times` int not null default 0 COMMENT '已经重试次数',
    `max_retry_times` int not null default 3 COMMENT '最大重试次数',
    `fail_reason` text COMMENT '执行失败的信息',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除标记',
    PRIMARY KEY (`id`),
    index idx_task_type (`task_type`),
    index idx_status_created_at (`status`, `created_at`),
    index idx_deleted (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本地消息表';