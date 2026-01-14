// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"net/http"

	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/logic/knowledge_document"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传文档
func UploadDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadDocumentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 解析多文件上传 (限制最大 10 个文件)
		err := r.ParseMultipartForm(100 << 20) // 100MB 最大内存
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			httpx.ErrorCtx(r.Context(), w, xerr.NewErrCodeMsg(xerr.BadRequest, "至少上传一个文件"))
			return
		}

		if len(files) > 10 {
			httpx.ErrorCtx(r.Context(), w, xerr.NewErrCodeMsg(xerr.BadRequest, "单次最多上传10个文件"))
			return
		}

		l := knowledge_document.NewUploadDocumentLogic(r.Context(), svcCtx)
		resp, err := l.UploadDocuments(&req, files)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
