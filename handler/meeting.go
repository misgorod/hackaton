package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/misgorod/hackaton/common"
	"github.com/misgorod/hackaton/model"
	"net/http"
)

type Meeting struct {
	Db *sql.DB
}

type meetingPostRequest struct {
	Amount float64
}

func (p *Meeting) Post(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var reqBody *meetingPostRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	meeting := model.Meeting{
		Id:           "",
		OwnerId:      id,
		Amount:       reqBody.Amount,
		Status:       "0",
		Participants: nil,
	}
	result, err := p.Db.ExecContext(r.Context(), "insert into public.event (owner, amount, state) values($1, $2, $3)", meeting.OwnerId, meeting.Amount, meeting.Status)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondJSON(w, 200, insertedId)
}

func (p *Meeting) GetAll(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rows, err := p.Db.QueryContext(r.Context(), "select e.id, e.amount from public.event e where e.owner = $1", id)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	meetings := make([]model.Meeting, 0)
	for rows.Next() {
		var meeting model.Meeting
		err := rows.Scan(&meeting.Id, &meeting.Amount)
		if err != nil {
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
			return
		}
		meeting.Status = "0"
		meeting.OwnerId = id
		meetings = append(meetings, meeting)
	}
	common.RespondJSON(w, http.StatusOK, meetings)
}

type meetingPutRequest struct {
	Amount  float64
	Invoice string
}

func (p *Meeting) Put(w http.ResponseWriter, r *http.Request) {
	ownerId := chi.URLParam(r, "ownerId")
	meetingId := chi.URLParam(r, "meetingId")

	var reqBody *meetingPutRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}

	result, err := p.Db.ExecContext(r.Context(), "insert into public.participant	(id_event, id_user, amount, invoice, state)	values ($1, $2, $3, $4, $5)", meetingId, ownerId, reqBody.Amount, reqBody.Invoice, "0")
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondJSON(w, 200, insertedId)
}
