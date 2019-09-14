package handler

import (
	"database/sql"
	"fmt"
	"github.com/misgorod/hackaton/common"
	"net/http"
)

type Invoice struct {
	db *sql.DB
}

func (i *Invoice) Post(w http.ResponseWriter, r *http.Request) {
	var id int
	err := i.db.QueryRowContext(r.Context(), "SELECT nextval('public.serial')").Scan(&id)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
	}
	common.RespondJSON(w, 200, id)
}