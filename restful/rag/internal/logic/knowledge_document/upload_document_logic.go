package knowledge_document

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"gozero-rag/internal/model/knowledge_base"
	"gozero-rag/internal/model/knowledge_document"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"github.com/google/uuid"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传文档
func NewUploadDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadDocumentLogic {
	return &UploadDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UploadDocuments 批量上传文档
func (l *UploadDocumentLogic) UploadDocuments(
	req *types.UploadDocumentReq,
	headers []*multipart.FileHeader,
) (resp *types.UploadDocumentResp, err error) {

	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. 验证知识库权限
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, req.KnowledgeBaseId)
	if err != nil {
		if err == knowledge_base.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "知识库不存在")
		}
		return nil, xerr.NewInternalErrMsg(err.Error())
	}
	if kb.TenantId != tenantId {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权操作此知识库")
	}

	// 2. 批量上传文件
	var fileIds []string
	var files []types.UploadedFileInfo
	var uploadErrors []string

	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("%s: 打开失败", header.Filename))
			continue
		}

		// 上传单个文件
		docId, docName, err := l.uploadSingleFile(
			req.KnowledgeBaseId,
			tenantId,
			userId,
			file,
			header,
		)
		file.Close()

		if err != nil {
			logx.Errorf("Upload file %s failed: %v", header.Filename, err)
			uploadErrors = append(uploadErrors, fmt.Sprintf("%s: %v", header.Filename, err))
			continue
		}

		fileIds = append(fileIds, docId)
		files = append(files, types.UploadedFileInfo{
			Id:      docId,
			DocName: docName,
		})
	}

	// 3. 返回结果
	if len(files) == 0 {
		return nil, xerr.NewErrCodeMsg(
			xerr.KnowledgeDocUploadError,
			fmt.Sprintf("所有文件上传失败: %s", uploadErrors),
		)
	}

	return &types.UploadDocumentResp{
		FileIds: fileIds,
		Files:   files,
	}, nil
}

// uploadSingleFile 上传单个文件 (提取出来的辅助函数)
func (l *UploadDocumentLogic) uploadSingleFile(
	knowledgeBaseId, tenantId, userId string,
	file multipart.File,
	header *multipart.FileHeader,
) (docId, docName string, err error) {

	// UUID 生成
	docUuid, err := uuid.NewV7()
	if err != nil {
		return "", "", xerr.NewInternalErrMsg("UUID generation failed")
	}
	docId = docUuid.String()

	ext := filepath.Ext(header.Filename)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// OSS 上传
	objectKey := fmt.Sprintf("tenant_%s/kb_%s/%s%s", tenantId, knowledgeBaseId, docId, ext)
	_, err = l.svcCtx.OssClient.PutObject(
		l.ctx,
		l.svcCtx.Config.Oss.BucketName,
		objectKey,
		file,
		header.Size,
		contentType,
	)
	if err != nil {
		return "", "", xerr.NewErrCodeMsg(xerr.KnowledgeDocUploadError, "OSS上传失败")
	}

	// 获取知识库配置以继承解析规则
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, knowledgeBaseId)
	if err != nil {
		logx.Errorf("UploadSingleFile: failed to find knowledge base %s: %v", knowledgeBaseId, err)
		return "", "", xerr.NewErrCodeMsg(xerr.KnowledgeBaseNotFoundError, "获取知识库信息失败")
	}

	// 数据库保存
	now := time.Now()
	docType := "unknown"
	if len(ext) > 1 {
		docType = ext[1:] // 去掉点号
	}

	// 处理 ParserConfig
	parserConfig := "{}"
	if kb.ParserConfig.Valid && kb.ParserConfig.String != "" {
		parserConfig = kb.ParserConfig.String
	}

	doc := &knowledge_document.KnowledgeDocument{
		Id:              docId,
		KnowledgeBaseId: knowledgeBaseId,
		DocName:         sql.NullString{String: header.Filename, Valid: true},
		DocType:         docType,
		DocSize:         int64(header.Size),
		StoragePath:     sql.NullString{String: objectKey, Valid: true},
		Status:          1,
		RunStatus:       "pending",
		CreatedBy:       userId,
		CreatedTime:     now.UnixMilli(),
		UpdatedTime:     now.UnixMilli(),
		CreatedDate:     now,
		UpdatedDate:     now,
		ParserId:        kb.ParserId,
		ParserConfig:    parserConfig,
		SourceType:      "local",
		Progress:        0,
		TokenNum:        0,
		ChunkNum:        0,
	}

	_, err = l.svcCtx.KnowledgeDocumentModel.Insert(l.ctx, doc)
	if err != nil {
		// 回滚 OSS
		_ = l.svcCtx.OssClient.RemoveObject(l.ctx, l.svcCtx.Config.Oss.BucketName, objectKey)
		return "", "", xerr.NewInternalErrMsg("保存文档信息失败")
	}

	// TODO: 发送异步解析任务到 Kafka/Asynq

	return docId, header.Filename, nil
}
