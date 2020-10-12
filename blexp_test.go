package blexp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/lukasmalkmus/expensify-go"
	"github.com/stretchr/testify/assert"
)

func reader() io.Reader {
	var data = `
	UserID = "WzSR7CgIa"
	UserSecret = "PYlLneWtPBaoIKtObILX1Y5jwQUMaTyh5ARf9klPe"
	UserEmail = "me@example.com"
	Default = "Broadband"

	[Templates]
	[Templates.Whatever]
	Merchant = "Somebody"
	Amount = 1122
	Currency = "USD"
	Category = "Entertainment"
	
	[Templates.Broadband]
	Merchant = "Xfinity"
	Amount = 2500
	Currency = "EUR"
	Category = "Employee Reimbursement"`
	return bytes.NewReader([]byte(data))
}

func readerNoTemplates() io.Reader {
	var data = `
	UserID = "WzSR7CgIa"
	UserSecret = "PYlLneWtPBaoIKtObILX1Y5jwQUMaTyh5ARf9klPe"
	UserEmail = "me@example.com"
	Default = "Broadband"`
	return bytes.NewReader([]byte(data))
}

type TestReader struct{}

func (r *TestReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("failed to read")
}

type TestTransport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.fn(req)
}

func (t *TestTransport) ReturnError(req *http.Request) (*http.Response, error) {
	return nil, errors.New("ReturnError")
}

func (t *TestTransport) SubmittedExpenses(n int) func(*http.Request) (*http.Response, error) {
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

func Test_ReadTOML(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	b, err := New(reader())
	a.NoError(err)
	a.NotNil(b.config)

	a.Equal("me@example.com", b.config.UserEmail)
	a.Equal(2, len(b.config.Templates))
	a.Equal("Employee Reimbursement", b.config.Templates["Broadband"].Category)
	a.Equal(2500, b.config.Templates["Broadband"].Amount)
	a.Equal("Broadband", b.Default())
	a.Equal(2, len(b.Templates()))

	b, err = New(&TestReader{})
	a.Error(err)
	a.Nil(b)
	a.Equal("failed to read", err.Error())

	b, err = New(readerNoTemplates())
	a.Error(err)
	a.Nil(b)
}

func Test_SubmitExpense(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	transport := &TestTransport{}
	c, err := expensify.NewClient("foo", "bar", expensify.SetClient(&http.Client{Transport: transport}))
	a.NotNil(c)
	a.NoError(err)

	b, err := New(reader())
	b.SetClient(c)
	a.NoError(err)

	ctx := context.Background()

	// expensify failed
	transport.fn = transport.ReturnError
	txn, err := b.SubmitExpense(ctx, "Broadband")
	a.Nil(txn)
	a.Error(err)
	a.Equal(`Post "https://integrations.expensify.com/Integration-Server/ExpensifyIntegrations": ReturnError`, err.Error())

	// success
	transport.fn = transport.SubmittedExpenses(1)
	txn, err = b.SubmitExpense(ctx, "Broadband")
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// expense template doesn't exist
	transport.fn = transport.SubmittedExpenses(1)
	txn, err = b.SubmitExpense(ctx, "foobar")
	a.Nil(txn)
	a.Error(err)
	a.Equal("failed to find expense template for name {foobar}", err.Error())

	// success default
	transport.fn = transport.SubmittedExpenses(1)
	txn, err = b.SubmitDefault(ctx)
	a.NotNil(txn)
	a.NoError(err)
	a.Equal("6720309558248016", txn.TransactionID)

	// failure no responses
	transport.fn = transport.SubmittedExpenses(0)
	txn, err = b.SubmitDefault(ctx)
	a.Nil(txn)
	a.Error(err)

	// failure too many responses
	transport.fn = transport.SubmittedExpenses(2)
	txn, err = b.SubmitDefault(ctx)
	a.Nil(txn)
	a.Error(err)
}
