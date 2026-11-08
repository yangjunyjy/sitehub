package handler

import (
	"mime"
	"sitehub/common/rpcx"
	"sitehub/common/streamio"
	"sitehub/service/blog/api/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func readAssetHandler(s *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stream, err := s.RPC.ReadAsset(rpcx.Outgoing(r.Context(), s.Config.ServiceKey), &pb.AssetRequest{Id: r.URL.Query().Get("id")})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// A bounded disk spool permits seekable ranges for object storage as well as local files.
		file, meta, err := streamio.Receive(r.Context(), stream.Recv, 50<<20)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer streamio.Remove(file)
		w.Header().Set("Content-Type", meta.Mime)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": meta.Name}))
		http.ServeContent(w, r, meta.Name, time.Time{}, file)
	}
}
