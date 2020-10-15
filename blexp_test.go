package blexp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/lukasmalkmus/expensify-go"
	"github.com/stretchr/testify/assert"
)

var templates = map[string]expensify.Expense{
	"Whatever":  expensify.Expense{},
	"Broadband": expensify.Expense{},
}

func ReturnError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("ReturnError")
}

func SubmittedExpenses(n int) func(*http.Request) (*http.Response, error) {
	exps := make([]string, n)
	for i := 0; i < n; i++ {
		exps[i] = `{
				"amount" : 2500,
				"merchant" : "Xfinity",
				"created" : "2016-01-01",
				"transactionID" : "6720309558248016",
				"currency" : "EUR"
			}`
	}
	return func(req *http.Request) (*http.Response, error) {
		var submittedJSON = fmt.Sprintf(
			`{"responseCode" : 200, "transactionList" : [%s]}`,
			strings.Join(exps, ","))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString(submittedJSON)),
			Header:     make(http.Header),
		}, nil
	}
}

func Test_New(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	b, err := New("user0192", "Cq355CzTQZCp",
		WithTemplates(templates, "Broadband"),
		WithUserEmail("me@example.com"))
	a.NoError(err)
	a.NotNil(b)
	a.Equal("me@example.com", b.UserEmail)
	a.Equal("Broadband", b.Default)
	a.Equal(2, len(b.Templates))
}

func Test_WithTemplates(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	b := &Blexp{}
	err := WithTemplates(templates, "foobar")(b)
	a.Error(err)

	err = WithTemplates(make(map[string]expensify.Expense), "foobar")(b)
	a.Error(err)
}

func Test_SubmitExpense(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	ctx := context.Background()
	b, err := New("user9102", "Cq35RuyzTQZCp",
		WithTemplates(templates, "Broadband"),
		WithUserEmail("me@example.com"))
	a.NoError(err)
	a.NotNil(b)

	// expensify failed
	WithTransport(ReturnError)(b)
	txn, err := b.SubmitExpense(ctx, "Broadband")
	a.Nil(txn)
	a.Error(err)
	a.Equal(`Post "https://integrations.expensify.com/Integration-Server/ExpensifyIntegrations": ReturnError`, err.Error())

	// success
	WithTransport(SubmittedExpenses(1))(b)
	txn, err = b.SubmitExpense(ctx, "Broadband")
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// expense template doesn't exist
	WithTransport(SubmittedExpenses(1))(b)
	txn, err = b.SubmitExpense(ctx, "foobar")
	a.Nil(txn)
	a.Error(err)
	a.Equal("failed to find expense template for name {foobar}", err.Error())

	// success default
	WithTransport(SubmittedExpenses(1))(b)
	txn, err = b.SubmitExpense(ctx, b.Default)
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// failure no responses
	WithTransport(SubmittedExpenses(0))(b)
	txn, err = b.SubmitExpense(ctx, b.Default)
	a.Nil(txn)
	a.Error(err)

	// failure too many responses
	WithTransport(SubmittedExpenses(2))(b)
	txn, err = b.SubmitExpense(ctx, b.Default)
	a.Nil(txn)
	a.Error(err)
}
