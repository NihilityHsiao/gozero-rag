// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowledge_base

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-rag/restful/rag/internal/logic/knowledge_base"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
)

// 获取知识库列表
func ListKnowledgeBaseHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListKnowledgeBaseReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowledge_base.NewListKnowledgeBaseLogic(r.Context(), svcCtx)
		resp, err := l.ListKnowledgeBase(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
