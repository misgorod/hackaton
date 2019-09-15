package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/misgorod/hackaton/common"
	"github.com/misgorod/hackaton/model"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"strconv"
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
		state := "1"
		if len(recipients) == 0 {
			state = "0"
		} else {
			for _, recipient := range recipients {
				if recipient.State == "0" {
					state = "0"
					break
				}
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
	Name    string `validate:"required,gte=1"`
}

func (p *Meeting) Put(w http.ResponseWriter, r *http.Request) {
	meetingId := chi.URLParam(r, "meetingId")
	meetingIdInt, err := strconv.Atoi(meetingId)
	if err != nil {
		common.RespondError(w, http.StatusBadRequest, "meetingId must be integer")
		return
	}
	var reqBody *meetingPutRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	if err := p.Validate.Struct(reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Validating error: %v", err))
		return
	}
	invoiceInt, err := strconv.Atoi(reqBody.Invoice)
	if err != nil {
		common.RespondError(w, http.StatusBadRequest, "invoice must be int")
		return
	}
	err = p.createParticipant(r.Context(), meetingIdInt, reqBody.Name, reqBody.Amount, invoiceInt)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.RespondOK(w)
}

func (p *Meeting) createParticipant(ctx context.Context, meetingId int, name, amount string, invoice int) error {
	_, err := p.Db.ExecContext(ctx, "insert into public.participant (id_event, name, amount, invoice) values ($1, $2, $3, $4)", meetingId, name, amount, invoice)
	if err != nil {
		return err
	}
	return nil
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

type recipientPostRequest struct {
	Amount    string `validate:"required,gte=1"`
	Payer string `validate:"required,gte=1"`
	Name string `validate:"required,gte=1"`
}

func (p *Meeting) PostRecipient(w http.ResponseWriter, r *http.Request) {
	ownerId := chi.URLParam(r, "ownerId")
	meetingId := chi.URLParam(r, "meetingId")
	meetingIdInt, err := strconv.Atoi(meetingId)
	if err != nil {
		common.RespondError(w, http.StatusBadRequest, "meeting id must be int")
		return
	}
	var reqBody *recipientPostRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	if err := p.Validate.Struct(reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Validating error: %v", err))
		return
	}

	var invoice int
	err = p.Db.QueryRowContext(r.Context(), "SELECT nextval('public.serial')").Scan(&invoice)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	err = createOutInvoice(reqBody.Amount, strconv.Itoa(invoice), reqBody.Payer, ownerId)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = p.createParticipant(r.Context(), meetingIdInt, reqBody.Name, reqBody.Amount, invoice)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.RespondOK(w)
}

func (p *Meeting) getRecipientsStatus(ctx context.Context, meetingId string) ([]model.Participant, error) {
	rows, err := p.Db.QueryContext(ctx, "select p.name, p.amount, p.invoice from public.participant p where p.id_event = $1", meetingId)
	if err != nil {
		return nil, err
	}
	var participants []model.Participant = make([]model.Participant, 0)
	for rows.Next() {
		var participant model.Participant
		err := rows.Scan(&participant.UserName, &participant.Amount, &participant.Invoice)
		if err != nil {
			return nil, err
		}
		state, err := getStateInvoice(participant.Invoice, participant.UserName)
		if err != nil {
			return nil, err
		}
		participant.State = strconv.Itoa(state)
		participants = append(participants, participant)
	}
	return participants, nil
}

type stateInvoiceResponse struct {
	State int `json:"omitempty"`
}

func getStateInvoice(invoice string, recipient string) (int, error) {
	response, err := http.Get(fmt.Sprintf("http://89.208.84.235:31080/api/v1/invoice/810/%s/%s", invoice, recipient))
	if err != nil {
		return 0, err
	}
	var stateResponse *stateInvoiceResponse
	if json.NewDecoder(response.Body).Decode(&stateResponse); err != nil {
		return 0, err
	}
	if stateResponse.State == 1 {
		return 0, nil
	} else if stateResponse.State == 5 {
		return 1, nil
	} else {
		return 0, nil
	}
}

type createInvoiceRequest struct {
	Amount       float64 `json:"amount"`
	CurrencyCode int `json:"currencyCode"`
	Description  string `json:"description"`
	Number       string `json:"number"`
	Payer        string `json:"payer"'`
	Recipient    string `json:"recipient"`
}

func createOutInvoice(amount, invoice, payer, recipient string) error {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("Eror converting amount")
	}
	body := createInvoiceRequest{
		amountFloat,
		810,
		"Description",
		invoice,
		payer,
		recipient,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	response, err := http.Post("http://89.208.84.235:31080/api/v1/invoice", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(response.Body)
		return errors.New(fmt.Sprintf("Error while creating invoice: bodySent: %v : status code: %v : bodyRecieved: %v", bodyBytes, response.StatusCode, responseBody))
	}
	return nil
}
