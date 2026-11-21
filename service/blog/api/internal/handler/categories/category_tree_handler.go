// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package categories

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"sitehub/service/blog/api/internal/logic/categories"
	"sitehub/service/blog/api/internal/svc"
)

func CategoryTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := categories.NewCategoryTreeLogic(r.Context(), svcCtx)
		resp, err := l.CategoryTree()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
