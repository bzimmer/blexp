package blexp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/lukasmalkmus/expensify-go"
)

// Blexp holds the necessary bits
type Blexp struct {
	UserEmail string
	Primary   string
	Templates map[string]expensify.Expense

	client *expensify.Client
}

// Option .
type Option func(b *Blexp) error

type transport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.fn(req)
}

// New .
func New(userID, userSecret string, options ...Option) (*Blexp, error) {
	c, err := expensify.NewClient(userID, userSecret)
	if err != nil {
		return nil, err
	}

	b := &Blexp{client: c}
	for _, opt := range options {
		err := opt(b)
		if err != nil {
			return nil, err
		}
	}
	return b, err
}

// WithUserEmail .
func WithUserEmail(userEmail string) Option {
	return func(b *Blexp) error {
		b.UserEmail = userEmail
		return nil
	}
}

// WithTemplates .
func WithTemplates(templates map[string]expensify.Expense, defaults ...string) Option {
	return func(b *Blexp) error {
		if len(templates) == 0 {
			return errors.New("no templates found")
		}
		if len(defaults) > 0 {
			b.Primary = defaults[0]
			if len(defaults) > 1 {
				log.Warn().Msg("using the first arg as primary, the rest are ignored")
			}
		}
		if _, ok := templates[b.Primary]; !ok {
			return fmt.Errorf("primary template {%s} not found", b.Primary)
		}
		b.Templates = templates
		return nil
	}
}

// WithTransport .
func WithTransport(f func(*http.Request) (*http.Response, error)) Option {
	return func(b *Blexp) error {
		b.client.Options(
			expensify.SetClient(
				&http.Client{Transport: &transport{f}},
			),
		)
		return nil
	}
}

// PrepareExpense creates an expense from name ready for submission
func (b *Blexp) PrepareExpense(name string) (*expensify.Expense, error) {
	exp, ok := b.Templates[name]
	if !ok {
		return nil, fmt.Errorf("failed to find expense template for name {%s}", name)
	}
	exp.Created = expensify.NewTime(time.Now())
	exp.Comment = fmt.Sprintf("blexp: %s", strings.Split(uuid.New().String(), "-")[0])
	return &exp, nil
}

// SubmitExpense submits the expense specified by name
func (b *Blexp) SubmitExpense(ctx context.Context, exp *expensify.Expense) (*expensify.SubmittedExpense, error) {
	log.Info().Str("op", "submit").Interface("exp", exp).Msg("submitting expense")
	submitted, err := b.client.Expense.Create(ctx, b.UserEmail, []*expensify.Expense{exp})
	if err != nil {
		return nil, err
	}
	n := len(submitted)
	if n != 1 {
		return nil, fmt.Errorf("expected one SubmittedExpense, received %d", n)
	}
	return submitted[0], err
}
