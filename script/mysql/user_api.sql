use gozero_rag;
drop table if exists user_api;
CREATE TABLE `user_api` (
   `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
   `user_id` varchar(32) NOT NULL COMMENT '用户ID (UUID)',
   `config_name` varchar(64) NOT NULL COMMENT '配置名称（如：通用问答模型、长文本总结模型）',
   `api_key` varchar(256) NOT NULL COMMENT 'open ai API密钥',
   `base_url` varchar(512) NOT NULL DEFAULT 'https://api.siliconflow.cn/v1' COMMENT '基础请求地址',
   `model_name` varchar(128) NOT NULL COMMENT '模型名称（如：deepseek-chat、qwen-7b-chat）',
   `model_type` varchar(32) NOT NULL COMMENT '模型类型（如：chat、embedding、qa、rewrite、rerank）',
   `model_dim` int NOT NULL DEFAULT 768 COMMENT '模型向量维度',
   `max_tokens` int NOT NULL DEFAULT 2048 COMMENT '单次请求最大tokens数',
   `temperature` float NOT NULL DEFAULT 0.7 COMMENT '温度系数（0-1，控制生成随机性）',
   `top_p` float NOT NULL DEFAULT 0.95 COMMENT '采样Top-P值',
   `timeout` int NOT NULL DEFAULT 30 COMMENT '请求超时时间（秒）',
   `status` tinyint NOT NULL DEFAULT 1 COMMENT '配置状态：0-禁用，1-启用',
   `is_default` tinyint(1) DEFAULT 0 COMMENT '是否为默认配置',
   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   PRIMARY KEY (`id`) USING BTREE,
   UNIQUE KEY `uk_uid_model_type_name` (`user_id`,`model_type`,`model_name`) USING BTREE COMMENT '模型名称唯一（单个用户单模型单配置）',
   KEY `idx_userid` (`user_id`) comment 'user表id',
   KEY `idx_status` (`status`) USING BTREE COMMENT '按状态筛选启用/禁用的配置',
   KEY `idx_model_type` (`model_type`) USING BTREE COMMENT '按模型类型筛选（如只查embedding模型）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='RAG-用户API配置表';