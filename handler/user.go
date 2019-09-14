package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/misgorod/hackaton/common"
	"github.com/misgorod/hackaton/model"
	"net/http"
)

type User struct {
	Db *sql.DB
}

func (p *User) Post(w http.ResponseWriter, r *http.Request) {
	var user *model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	var id string
	err := p.Db.QueryRowContext(r.Context(), "select id from public.user where id = $1", user.Id).Scan(&id)
	if err == nil {
		common.RespondJSON(w, http.StatusBadRequest, "user already exists")
		return
	} else if err != sql.ErrNoRows {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	if _, err := p.Db.ExecContext(r.Context(), "insert into public.user(id, name) values($1, $2)", user.Id, ""); err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondJSON(w, 200, id)
}
