use gozero_rag;
drop table if exists `knowledge_base`;
CREATE TABLE `knowledge_base`
(
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` varchar(64) NOT NULL COMMENT '知识库名称',
    `description` varchar(256) DEFAULT NULL COMMENT '知识库描述',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '知识库状态：0-禁用，1-启用',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `embedding_model_id` bigint unsigned not null default 0 comment '强绑定 user_api表的embedding模型id,不可修改',
    `model_ids` json default (JSON_OBJECT()) comment '弱绑定, qa/chat/rewrite/rerank模型的id,可修改',
    primary key (`id`),
    unique key `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='RAG-知识库核心表';


insert into `knowledge_base`(`name`, `description`) values ('默认知识库', '默认知识库');


drop table if exists `knowledge_document`;
CREATE TABLE `knowledge_document`
(
    `id`           char(36) NOT NULL COMMENT 'uuid v7 主键ID',
    `knowledge_base_id` bigint unsigned NOT NULL COMMENT '知识库ID',
    `doc_name`         varchar(256)     NOT NULL COMMENT '文档名称',
    `doc_type`         varchar(32)     NOT NULL COMMENT '文档类型(pdf/word/txt)',
    `storage_path` varchar(512) not null comment '文档存储路径',
    `doc_size` bigint not null comment '文档大小（字节）',
    `description`  varchar(256)             DEFAULT NULL COMMENT '文档描述',
    `status`       varchar(32)         NOT NULL DEFAULT 'pending' COMMENT '文档状态, disable-禁用, pending-待处理,indexing-索引中, enable-可用, fail-处理失败',
    `chunk_count` int not null default 0 comment '文档分片数量',
    `err_msg` varchar(512) not null default '' comment '失败原因',
    `parser_config` json not null comment '分词与清洗规则配置',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    primary key (`id`),
    index `idx_knowledge_base_id` (`knowledge_base_id`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT ='RAG-知识库文档表';


drop table if exists `knowledge_document_chunk`;
CREATE TABLE `knowledge_document_chunk`
(
    `id`                    char(36) NOT NULL COMMENT '分片id-uuid-v7',
    `knowledge_base_id`     bigint unsigned NOT NULL COMMENT '知识库ID',
    `knowledge_document_id` char(36) NOT NULL COMMENT '文档ID',
    `chunk_text`            text            NOT NULL COMMENT '分片内容',
    `chunk_size`            int             not null comment '分片大小（字节）',
    `metadata`              json            default (JSON_OBJECT())  comment '分片元数据',
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `status`       varchar(32)         NOT NULL DEFAULT 'disable' COMMENT 'disable-禁用, enable-启用, qa-正在执行qa生成, embedding-正在生成向量',
    primary key (`id`),
    key `idx_kdoc_id` (`knowledge_document_id`),
    key `idx_kbase_id` (`knowledge_base_id`)
)ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT ='RAG-知识库文档分片表';
