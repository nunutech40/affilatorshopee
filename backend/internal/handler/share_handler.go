package handler

import (
	"net/http"

	"github.com/nunutech40/affilatorshopee/internal/service"
)

type ShareHandler struct{ share *service.ShareService }

func NewShareHandler(share *service.ShareService) *ShareHandler { return &ShareHandler{share: share} }

func (h *ShareHandler) X(w http.ResponseWriter, r *http.Request) {
	caption := r.URL.Query().Get("caption")
	if caption == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "caption wajib diisi")
		return
	}
	http.Redirect(w, r, h.share.TwitterIntentURL(caption, r.URL.Query()["media"]), http.StatusFound)
}
