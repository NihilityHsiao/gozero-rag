// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package team

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-rag/restful/rag/internal/logic/team"
	"gozero-rag/restful/rag/internal/svc"
)

// 获取当前团队成员列表
func ListMembersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := team.NewListMembersLogic(r.Context(), svcCtx)
		resp, err := l.ListMembers()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
