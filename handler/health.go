package handler

import (
	"github.com/misgorod/hackaton/common"
	"net/http"
)

type Health struct{}

func (h *Health) Get(w http.ResponseWriter, r *http.Request) {
	common.RespondJSON(w, 200, "OK")
}
