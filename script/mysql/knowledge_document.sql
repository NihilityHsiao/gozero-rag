use gozero_rag;

DROP TABLE IF EXISTS `knowledge_document`;

CREATE TABLE `knowledge_document` (
  `id` char(36) NOT NULL COMMENT 'UUID v7',
  `knowledge_base_id` varchar(36) NOT NULL COMMENT '所属知识库ID',
  `doc_name` varchar(255) DEFAULT NULL COMMENT '文档名称',
  `doc_type` varchar(32) NOT NULL COMMENT '类型: pdf/word/txt/md',
  `doc_size` int(11) NOT NULL DEFAULT 0 COMMENT '大小(字节)',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  `storage_path` varchar(255) DEFAULT NULL COMMENT '存储路径(MinIO)',
  `source_type` varchar(128) NOT NULL DEFAULT 'local' COMMENT '来源: local/url',
  `created_by` varchar(36) NOT NULL COMMENT '上传者ID',
  `token_num` int(11) NOT NULL DEFAULT 0 COMMENT 'Token数',
  `chunk_num` int(11) NOT NULL DEFAULT 0 COMMENT '分片数',
  `progress` float NOT NULL DEFAULT 0 COMMENT '处理进度(0-1)',
  `progress_msg` longtext COMMENT '进度/错误信息',
  `process_begin_at` datetime DEFAULT NULL COMMENT '开始处理时间',
  `process_duration` float NOT NULL DEFAULT 0 COMMENT '处理耗时(s)',
  `run_status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/running/success/fail/paused',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1-有效, 0-删除/禁用',

  `parser_id` varchar(36) NOT NULL DEFAULT 'general' COMMENT '解析器ID,目前仅支持 general | resume',
  `parser_config` longtext NOT NULL COMMENT '解析配置(JSON)',


  `created_time` bigint NOT NULL COMMENT '创建时间戳(ms)',
  `updated_time` bigint NOT NULL COMMENT '更新时间戳(ms)',
  `created_date` datetime NOT NULL COMMENT '创建日期',
  `updated_date` datetime NOT NULL COMMENT '更新日期',
  `meta_fields` longtext COMMENT '元数据',
  PRIMARY KEY (`id`),
  KEY `idx_doc_kb_id` (`knowledge_base_id`),
  KEY `idx_doc_name` (`doc_name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT ='文档表';