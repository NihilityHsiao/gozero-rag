// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_document

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-rag/restful/rag/internal/logic/knowledge_document"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
)

// 获取文档详情
func GetKnowledgeDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetKnowledgeDocumentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowledge_document.NewGetKnowledgeDocumentLogic(r.Context(), svcCtx)
		resp, err := l.GetKnowledgeDocument(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
