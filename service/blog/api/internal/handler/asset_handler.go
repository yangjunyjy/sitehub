package handler

import (
	"io"
	"mime"
	"net/http"
	"os"
	"sitehub/common/rpcx"
	"sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/domain"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func readAssetHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var actor domain.ArticleActor
		if rpcx.HTTPValue(r.Context(), "authorization") != "" {
			principal, err := s.Require(r.Context(), "", "admin", false)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			actor = logic.ArticleActor(principal)
		}
		asset, err := s.Core.Assets.ReadArticleAsset(r.Context(), actor, r.URL.Query().Get("id"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, logic.Classify(err))
			return
		}
		defer asset.Body.Close()
		if asset.Size <= 0 || asset.Size > 50<<20 {
			http.Error(w, "invalid stored asset size", http.StatusInternalServerError)
			return
		}
		var reader io.ReadSeeker
		if seekable, ok := asset.Body.(io.ReadSeeker); ok {
			reader = seekable
		} else {
			file, err := os.CreateTemp("", "sitehub-asset-*")
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
			defer func() { _ = file.Close(); _ = os.Remove(file.Name()) }()
			copied, err := io.Copy(file, io.LimitReader(asset.Body, asset.Size+1))
			if err != nil || copied != asset.Size {
				http.Error(w, "incomplete asset", http.StatusInternalServerError)
				return
			}
			reader = file
		}
		w.Header().Set("Content-Type", asset.MIME)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": asset.Name}))
		http.ServeContent(w, r, asset.Name, time.Time{}, reader)
	}
}
