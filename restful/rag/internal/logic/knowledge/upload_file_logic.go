// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gozero-rag/restful/rag/internal/common"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/metric"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxFileSize    = 50 << 20              // 50MB
	uploadBasePath = "./uploads/knowledge" // 文件存储基础路径
)

// 支持的文件类型
var supportedDocTypes = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".docx": true,
	".doc":  true,
	".md":   true,
}

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

// NewUploadFileLogic 上传 pdf/txt/docx 文件到指定知识库，使用 multipart/form-data，文件字段名: file
func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *UploadFileLogic) UploadFile(req *types.UploadFileReq) (resp *types.UploadFileResp, err error) {
	start := time.Now()
	_, err = common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	docType := "unknown" // 默认类型,用于失败时的指标记录

	// 1. 验证知识库是否存在
	kb, err := l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, uint64(req.KnowledgeBaseId))
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			metric.RecordUploadFail(docType, metric.FailReasonKbNotFound)
			return nil, xerr.NewErrCode(xerr.KnowledgeBaseNotFoundError)
		}
		l.Errorf("查询知识库失败: %v, knowledgeBaseId: %d", err, req.KnowledgeBaseId)
		metric.RecordUploadFail(docType, metric.FailReasonDbError)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 检查知识库状态
	if kb.Status == 0 {
		metric.RecordUploadFail(docType, metric.FailReasonKbDisabled)
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "知识库已禁用,无法上传文档")
	}

	// 2. 解析form表单,获取文件
	if err := l.r.ParseMultipartForm(maxFileSize); err != nil {
		l.Errorf("解析form表单失败: %v", err)
		metric.RecordUploadFail(docType, metric.FailReasonParseError)
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "文件大小超过限制或表单格式错误")
	}

	file, header, err := l.r.FormFile("file")
	if err != nil {
		l.Errorf("获取上传文件失败: %v", err)
		metric.RecordUploadFail(docType, metric.FailReasonParseError)
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "请选择要上传的文件")
	}
	defer file.Close()

	// 3. 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	docType = strings.TrimPrefix(ext, ".") // 更新docType用于指标
	if !supportedDocTypes[ext] {
		metric.RecordUploadFail(docType, metric.FailReasonInvalidType)
		return nil, xerr.NewErrCode(xerr.KnowledgeDocTypeNotSupport)
	}

	// 4. 验证文件大小
	if header.Size > maxFileSize {
		metric.RecordUploadFail(docType, metric.FailReasonFileTooLarge)
		return nil, xerr.NewErrCode(xerr.KnowledgeDocTooLargeError)
	}

	// 5. 生成存储路径并保存文件
	// 路径格式: ./uploads/knowledge/{knowledge_base_id}/{timestamp}_{filename}
	storagePath, err := l.saveFile(file, req.KnowledgeBaseId, header.Filename)
	if err != nil {
		l.Errorf("保存文件失败: %v", err)
		metric.RecordUploadFail(docType, metric.FailReasonSaveError)
		return nil, xerr.NewErrCode(xerr.KnowledgeDocSaveError)
	}

	// 6. 创建文档记录到数据库
	doc := &knowledge.KnowledgeDocument{
		KnowledgeBaseId: uint64(req.KnowledgeBaseId),
		DocName:         header.Filename,
		DocType:         docType,
		StoragePath:     storagePath,
		DocSize:         header.Size,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:          knowledge.StatusDocumentPending, // 默认启用
		ChunkCount:      0,                               // 初始分片数为0,后续异步处理
		ErrMsg:          "",
	}

	result, err := l.svcCtx.KnowledgeDocumentModel.Insert(l.ctx, doc)
	if err != nil {
		l.Errorf("插入文档记录失败: %v", err)
		// 回滚: 删除已保存的文件
		_ = os.Remove(storagePath)
		metric.RecordUploadFail(docType, metric.FailReasonDbError)
		return nil, xerr.NewErrCode(xerr.KnowledgeDocUploadError)
	}

	docId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取文档ID失败: %v", err)
		metric.RecordUploadFail(docType, metric.FailReasonDbError)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 8. 记录成功指标
	duration := time.Since(start).Seconds()

	metric.RecordUploadSuccess(docType, header.Size, duration)

	// 9. 返回响应
	return &types.UploadFileResp{
		Id:      docId,
		DocName: header.Filename,
		DocType: docType,
		DocSize: header.Size,
	}, nil
}

// saveFile 保存文件到本地存储
func (l *UploadFileLogic) saveFile(file io.Reader, knowledgeBaseId uint64, filename string) (string, error) {
	// 创建目录
	dir := fmt.Sprintf("%s/%d", uploadBasePath, knowledgeBaseId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成文件名: 时间戳_原文件名
	newFilename := fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filename)
	filePath := filepath.Join(dir, newFilename)

	// 创建文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return filePath, nil
}
