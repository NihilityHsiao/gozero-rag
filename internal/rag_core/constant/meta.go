package constant

// doc相关的元信息字段

// ========================
// Document Metadata 字段定义
// 用于 schema.Document.MetaData 的统一 key
// ========================

const (
	// --- 结构信息 (Structure Transformer 注入) ---

	// MetaHeaderContext 标题上下文 (如: "第一章 总则")
	MetaHeaderContext = "header_context"
	// MetaHeaderLevel 标题层级 (1: 章节, 2: 小节, ...)
	MetaHeaderLevel = "header_level"
	// MetaHeaderType 标题类型 (numbered/chinese/heuristic)
	MetaHeaderType = "header_type"

	// --- 来源信息 (Loader/Indexer 注入) ---

	// MetaSourceURI 原始文件路径或 URL
	MetaSourceURI = "source_uri"
	// MetaSourceDocID 原始文档 ID
	MetaSourceDocID = "source_doc_id"

	// --- 分片信息 ---

	// MetaChunkIndex 当前 chunk 在文档中的索引
	MetaChunkIndex = "chunk_index"
	// MetaChunkTotal 文档总 chunk 数 (可选，后处理填充)
	MetaChunkTotal = "chunk_total"

	// --- QA Checker 注入 (预留) ---

	// MetaQaPairs 生成的 QA 对 ([]QaPair)
	MetaQaPairs = "qa_pairs"
	// MetaSummary chunk 摘要
	MetaSummary = "summary"
	// MetaKeywords 关键词列表 ([]string)
	MetaKeywords = "keywords"
	// MetaQualityScore 质量评分 (0-100)
	MetaQualityScore = "quality_score"
	// MetaQualityDetails 详细评分结果 (*types.ChunkQualityScore)
	MetaQualityDetails = "quality_details"
)
