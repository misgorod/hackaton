package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/misgorod/hackaton/common"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Invoice struct {
	Db *sql.DB
}

func (i *Invoice) Post(w http.ResponseWriter, r *http.Request) {
	var id int
	err := i.Db.QueryRowContext(r.Context(), "SELECT nextval('public.serial')").Scan(&id)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	common.RespondJSON(w, 200, strconv.Itoa(id))
}

type invoicePutRequest struct {
	Amount    string
	Recipient string
}

func (i *Invoice) Put(w http.ResponseWriter, r *http.Request) {
	var reqBody *invoicePutRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		common.RespondError(w, http.StatusBadRequest, "cannot decode request")
		return
	}
	var invoice int
	err := i.Db.QueryRowContext(r.Context(), "SELECT nextval('public.serial')").Scan(&invoice)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
		return
	}
	err = createInvoice(reqBody.Amount, reqBody.Recipient, invoice)
	if err != nil {
		common.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("out error: %v", err))
		return
	}
	common.RespondJSON(w, 200, strconv.Itoa(invoice))
}

type createInvoiceRequest struct {
	Amount       float64 `json:"amount"`
	CurrencyCode int     `json:"currencyCode"`
	Description  string  `json:"description"`
	Number       string  `json:"number"`
	Recipient    string  `json:"recipient"`
}

func createInvoice(amount, recipient string, invoice int) error {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return errors.New("Eror converting amount")
	}
	body := createInvoiceRequest{
		amountFloat,
		810,
		"Description",
		strconv.Itoa(invoice),
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
