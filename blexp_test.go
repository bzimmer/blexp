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
	"Whatever":  {},
	"Broadband": {},
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
	a.Equal("Broadband", b.Primary)
	a.Equal(2, len(b.Templates))

	// test error in Option
	b, err = New("user0192", "Cq355CzTQZCp",
		func(b *Blexp) error {
			return fmt.Errorf("foobar")
		})
	a.Error(err)
	a.Nil(b)
}

func Test_WithTemplates(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	b := &Blexp{}
	err := WithTemplates(templates, "foobar")(b)
	a.Error(err)

	// test failure when default name is not found in templates
	err = WithTemplates(make(map[string]expensify.Expense), "foobar")(b)
	a.Error(err)

	// test success even if more than one key in list
	err = WithTemplates(templates, "Broadband", "baz", "bar")(b)
	a.NoError(err)
	a.Equal("Broadband", b.Primary)

	// test error when template key is in the list but not first
	err = WithTemplates(templates, "baz", "Broadband")(b)
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
	a.NoError(WithTransport(ReturnError)(b))
	exp, err := b.PrepareExpense("Broadband")
	a.NotNil(exp)
	a.NoError(err)
	txn, err := b.SubmitExpense(ctx, exp)
	a.Nil(txn)
	a.Error(err)
	a.Equal(`Post "https://integrations.expensify.com/Integration-Server/ExpensifyIntegrations": ReturnError`, err.Error())

	// success
	a.NoError(WithTransport(SubmittedExpenses(1))(b))
	exp, err = b.PrepareExpense("Broadband")
	a.NotNil(exp)
	a.NoError(err)
	txn, err = b.SubmitExpense(ctx, exp)
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// expense template doesn't exist
	exp, err = b.PrepareExpense("foobar")
	a.Nil(exp)
	a.Error(err)
	a.Equal("failed to find expense template for name {foobar}", err.Error())

	// success default
	a.NoError(WithTransport(SubmittedExpenses(1))(b))
	exp, err = b.PrepareExpense(b.Primary)
	a.NotNil(exp)
	a.NoError(err)
	txn, err = b.SubmitExpense(ctx, exp)
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// failure no responses
	a.NoError(WithTransport(SubmittedExpenses(0))(b))
	exp, err = b.PrepareExpense(b.Primary)
	a.NotNil(exp)
	a.NoError(err)
	txn, err = b.SubmitExpense(ctx, exp)
	a.Nil(txn)
	a.Error(err)

	// failure too many responses
	a.NoError(WithTransport(SubmittedExpenses(2))(b))
	exp, err = b.PrepareExpense(b.Primary)
	a.NotNil(exp)
	a.NoError(err)
	txn, err = b.SubmitExpense(ctx, exp)
	a.Nil(txn)
	a.Error(err)
}
