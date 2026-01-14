use gozero_rag;

drop table if exists `knowledge_base`;

CREATE TABLE `knowledge_base`
(
    `id` varchar(36) NOT NULL COMMENT '主键ID, UUID v7',
    `avatar` longtext COMMENT '知识库头像 base64',
    `tenant_id` varchar(36) NOT NULL COMMENT '所属租户ID, 核心隔离字段',
    `name` varchar(64) NOT NULL COMMENT '知识库名称',
    `language` varchar(32) NOT NULL DEFAULT 'Chinese' COMMENT '语言: English|Chinese, 影响分词策略',
    `description` varchar(256) DEFAULT NULL COMMENT '描述',
    `embd_id` varchar(128) NOT NULL COMMENT 'Embedding模型ID (来自租户配置)',
    `permission` varchar(16) NOT NULL DEFAULT 'me' COMMENT '权限: me(仅创建者)|team(租户全员)',
    `created_by` varchar(36) NOT NULL COMMENT '创建者用户ID',
    `doc_num` bigint unsigned NOT NULL DEFAULT 0 COMMENT '文档数量',
    `token_num` bigint unsigned NOT NULL DEFAULT 0 COMMENT 'Token总数',
    `chunk_num` bigint unsigned NOT NULL DEFAULT 0 COMMENT '分片总数',
    `similarity_threshold` float NOT NULL DEFAULT 0.3 COMMENT '检索相似度阈值',
    `vector_similarity_weight` float NOT NULL DEFAULT 0.3 COMMENT '混合检索向量权重 (1-keyword)',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',

    `parser_id` varchar(36) NOT NULL DEFAULT 'general' COMMENT '解析器ID,目前仅支持 general | resume',
    `parser_config` longtext COMMENT '解析器配置, 默认是 {}',

    `created_time` bigint NOT NULL COMMENT '创建时间戳(ms)',
    `updated_time` bigint NOT NULL COMMENT '更新时间戳(ms)',
    `created_date` datetime NOT NULL COMMENT '创建日期',
    `updated_date` datetime NOT NULL COMMENT '更新日期',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_tenant_id_name` (`tenant_id`, `name`),
    INDEX `idx_tenant_permission_status` (`tenant_id`, `permission`, `status`),
    INDEX `idx_created_by` (`created_by`),
    INDEX `idx_permission` (`permission`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库表';
