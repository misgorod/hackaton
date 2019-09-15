package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/misgorod/hackaton/common"
	"github.com/misgorod/hackaton/model"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type Meeting struct {
	Db       *sql.DB
	Validate *validator.Validate
}

type meetingPostRequest struct {
	Amount string `validate:"required,gte=1"`
	Name   string `validate:"required,gte=1"`
	Date   string `validate:"required,gte=1"`
}

func (p *Meeting) Post(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var reqBody *meetingPostRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	if err := p.Validate.Struct(reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Validating error: %v", err))
		return
	}
	var insertedId int
	err := p.Db.QueryRowContext(r.Context(), "insert into public.event (owner, amount, state, name, date) values($1, $2, $3, $4, $5) RETURNING id", id, reqBody.Amount, "0", reqBody.Name, reqBody.Date).Scan(&insertedId)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondJSON(w, 200, insertedId)
}

func (p *Meeting) GetAll(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rows, err := p.Db.QueryContext(r.Context(), "select e.id, e.amount, e.date, e.name from public.event e where e.owner = $1", id)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	meetings := make([]model.Meeting, 0)
	for rows.Next() {
		var meeting model.Meeting
		err := rows.Scan(&meeting.Id, &meeting.Amount, &meeting.Date, &meeting.Name)
		if err != nil {
			common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
			return
		}
		recipients, err := p.getRecipientsStatus(r.Context(), meeting.Id)
		if err != nil {
			common.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		state := "5"
		for _, recipient := range recipients {
			if recipient.State == "0" {
				state = "0"
				break
			}
		}
		meeting.State = state
		meeting.OwnerId = id
		meetings = append(meetings, meeting)
	}
	common.RespondJSON(w, http.StatusOK, meetings)
}

type meetingPutRequest struct {
	Amount  string `validate:"required,gte=1"`
	Invoice string `validate:"required,gte=1"`
}

func (p *Meeting) Put(w http.ResponseWriter, r *http.Request) {
	ownerId := chi.URLParam(r, "ownerId")
	meetingId := chi.URLParam(r, "meetingId")

	var reqBody *meetingPutRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	if err := p.Validate.Struct(reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Validating error: %v", err))
		return
	}
	_, err := p.Db.ExecContext(r.Context(), "insert into public.participant	(id_event, id_user, amount, invoice) values ($1, $2, $3, $4)", meetingId, ownerId, reqBody.Amount, reqBody.Invoice)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondOK(w)
}

func (p *Meeting) Get(w http.ResponseWriter, r *http.Request) {
	meetingId := chi.URLParam(r, "meetingId")
	participants, err := p.getRecipientsStatus(r.Context(), meetingId)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.RespondJSON(w, http.StatusOK, participants)
}

func (p *Meeting) getRecipientsStatus(ctx context.Context, meetingId string) ([]model.Participant, error) {
	rows, err := p.Db.QueryContext(ctx, "select p.id_user, u.name, p.amount, p.invoice from public.participant p join public.user u where p.id_event = $1", meetingId)
	if err != nil {
		return nil, err
	}
	var participants []model.Participant = make([]model.Participant, 0)
	for rows.Next() {
		var participant model.Participant
		err := rows.Scan(&participant.UserId, &participant.UserName, &participant.Amount, &participant.Invoice)
		if err != nil {
			return nil, err
		}
		state, err := GetStateInvoice(participant.Invoice, participant.UserId)
		if err != nil {
			return nil, err
		}
		participant.State = string(state)
		participants = append(participants, participant)
	}
	return participants, nil
}

type StateInvoiceResponse struct {
	State int `json:"omitempty"`
}

func GetStateInvoice(invoice string, recipient string) (int, error) {
	response, err := http.Get(fmt.Sprintf("http://89.208.84.235:31080/api/v1/invoice/810/%s/%s", invoice, recipient))
	if err != nil {
		return 0, err
	}
	var stateResponse *StateInvoiceResponse
	if json.NewDecoder(response.Body).Decode(&stateResponse); err != nil {
		return 0, err
	}
	if stateResponse.State == 1 {
		return 0, nil
	} else if stateResponse.State == 5 {
		return 1, nil
	} else {
		return 0, errors.New("Wrong state of invoice")
	}
}
