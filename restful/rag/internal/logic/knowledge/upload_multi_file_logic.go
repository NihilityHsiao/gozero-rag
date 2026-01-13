// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gozero-rag/internal/model/knowledge"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// 定义常量
const (
	MaxFileSize = 30 << 20 // 30MB
	UploadDir   = "./uploads"
)

// 允许的文件扩展名白名单
var allowedExts = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".md":   true,
	".docx": true,
	".doc":  true,
}

type UploadMultiFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

// NewUploadMultiFileLogic 上传 pdf/txt/docx 文件到指定知识库，使用 multipart/form-data，文件字段名: files
func NewUploadMultiFileLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UploadMultiFileLogic {
	return &UploadMultiFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *UploadMultiFileLogic) UploadMultiFile(req *types.UploadMultiFileReq) (resp *types.UploadMultiFileResp, err error) {
	// todo: add your logic here and delete this line
	_, err = common.GetUidFromCtx(l.ctx)

	if err != nil {
		return nil, err
	}

	kb, err := l.checkKnowledgeBase(req)
	if err != nil {
		return nil, err
	}

	// 参数是 maxMemory，例如 32MB。超过部分会存到系统临时文件，未超过部分在内存
	err = l.r.ParseMultipartForm(32 << 20)
	if err != nil {
		logx.Errorf("解析表单失败: %v", err)
		return nil, xerr.NewInternalErrMsg("解析表单失败")
	}

	if l.r.MultipartForm == nil || l.r.MultipartForm.File == nil {
		return nil, xerr.NewBadRequestErrMsg("未上传任何文件")
	}

	fileHeaders := l.r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		return nil, fmt.Errorf("文件列表为空")
	}

	// 返回的id列表
	uploadedIds := make([]string, 0)

	// filename 和 storagePath
	filenamePathMap := make(map[string]string)

	// 临时文件目录，用于可能的临时处理，或者直接流式上传
	// 这里直接流式上传，不需要本地目录
	// dir := fmt.Sprintf("%s/%d", uploadBasePath, kb.Id)

	for _, fileHeader := range fileHeaders {
		// 1.验证文件类型
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		//docType := strings.TrimPrefix(ext, ".") // 更新docType用于指标
		if !allowedExts[ext] {
			return nil, xerr.NewErrCode(xerr.KnowledgeDocTypeNotSupport)
		}

		// 2. 验证文件大小
		if fileHeader.Size > MaxFileSize {
			return nil, xerr.NewBadRequestErrMsg("文件过大：" + fileHeader.Filename)
		}

		_, exist := filenamePathMap[fileHeader.Filename]
		if exist {
			return nil, xerr.NewBadRequestErrMsg("重复文件名:" + fileHeader.Filename)
		}

		filename := fileHeader.Filename

		fileId, processErr := func() (string, error) {
			file, err := fileHeader.Open()
			if err != nil {
				return "", err
			}
			defer file.Close()

			// 生成uuid v7
			uuidObj, err := uuid.NewV7()
			if err != nil {
				return "", xerr.NewInternalErrMsg("生成uuid失败")
			}
			docId := uuidObj.String()
			objectKey := fmt.Sprintf("kb_%d/%s/%s", kb.Id, docId, filename)

			// 3. 上传到 MinIO
			_, err = l.svcCtx.OssClient.PutObject(l.ctx, l.svcCtx.Config.Oss.BucketName, objectKey, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
			if err != nil {
				return "", fmt.Errorf("上传OSS失败: %v", err)
			}

			// 记录 path 为 object key
			absPath := objectKey

			// 入库
			doc := &knowledge.KnowledgeDocument{
				Id:              docId,
				KnowledgeBaseId: kb.Id,
				DocName:         fileHeader.Filename,
				DocType:         strings.TrimPrefix(ext, "."),
				StoragePath:     absPath,
				DocSize:         fileHeader.Size,
				Description: sql.NullString{
					String: req.Description,
					Valid:  req.Description == "",
				},
				// 只有 save parser_config后，才会设置成pending
				Status:       knowledge.StatusDocumentDisable,
				ChunkCount:   0,
				ErrMsg:       "",
				ParserConfig: "{}",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			_, err = l.svcCtx.KnowledgeDocumentModel.Insert(l.ctx, doc)
			if err != nil {
				l.Errorf("插入文档记录失败: %v", err)
				// 回滚: 删除已上传的 MinIO 文件
				_ = l.svcCtx.OssClient.RemoveObject(l.ctx, l.svcCtx.Config.Oss.BucketName, objectKey)
				return "", xerr.NewErrCode(xerr.KnowledgeDocUploadError)
			}

			return docId, nil
		}()

		if processErr != nil {
			logx.Errorf("处理文件 %s 失败:%v", fileHeader.Filename, processErr)
		} else {
			uploadedIds = append(uploadedIds, fileId)
		}

		if len(uploadedIds) == 0 {
			return nil, xerr.NewInternalErrMsg("所有文件上传失败")
		}
	}

	return &types.UploadMultiFileResp{
		FileIds: uploadedIds,
	}, nil
}

func (l *UploadMultiFileLogic) checkKnowledgeBase(req *types.UploadMultiFileReq) (kb *knowledge.KnowledgeBase, err error) {
	kb, err = l.svcCtx.KnowledgeBaseModel.FindOne(l.ctx, uint64(req.KnowledgeBaseId))
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			// 知识库不存在
			return nil, xerr.NewErrCode(xerr.KnowledgeBaseNotFoundError)
		}
		l.Errorf("查询知识库失败: %v, knowledgeBaseId: %d", err, req.KnowledgeBaseId)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 检查知识库状态
	if kb.Status == 0 {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "知识库已禁用,无法上传文档")
	}

	return kb, nil

}
