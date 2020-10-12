package blexp

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/BurntSushi/toml"
	"github.com/lukasmalkmus/expensify-go"
)

// Config .
type Config struct {
	UserID     string
	UserSecret string
	UserEmail  string
	Default    string
	Templates  map[string]expensify.Expense
}

// Blexp holds the necessary bits
type Blexp struct {
	config *Config
	client *expensify.Client
}

// New creates a new blexp instance
func New(reader io.Reader) (*Blexp, error) {
	config := &Config{}
	_, err := toml.DecodeReader(reader, config)
	if err != nil {
		return nil, err
	}
	if config.Templates == nil {
		return nil, fmt.Errorf("no templates found")
	}

	client, err := expensify.NewClient(config.UserID, config.UserSecret)
	if err != nil {
		return nil, err
	}
	b := &Blexp{config: config, client: client}
	return b, nil
}

// SetClient sets the expensify client otherwise the default is used
func (b *Blexp) SetClient(c *expensify.Client) *Blexp {
	b.client = c
	return b
}

// Default returns the name of the default expense
func (b *Blexp) Default() string {
	return b.config.Default
}

// SubmitDefault submits the default expense
func (b *Blexp) SubmitDefault(ctx context.Context) (*expensify.SubmittedExpense, error) {
	return b.SubmitExpense(ctx, b.Default())
}

// Templates .
func (b *Blexp) Templates() map[string]expensify.Expense {
	return b.config.Templates
}

// SubmitExpense submits the expense specified by name
func (b *Blexp) SubmitExpense(ctx context.Context, name string) (*expensify.SubmittedExpense, error) {
	exp, ok := b.config.Templates[name]
	if !ok {
		return nil, fmt.Errorf("failed to find expense template for name {%s}", name)
	}
	expenses := []*expensify.Expense{&exp}
	exp.Created = expensify.NewTime(time.Now())
	exp.Comment = fmt.Sprintf("blexp: %s", strings.Split(uuid.New().String(), "-")[0])

	log.Info().Str("op", "submit").Interface("exp", exp).Msg("submitting expense")
	submitted, err := b.client.Expense.Create(ctx, b.config.UserEmail, expenses)
	if err != nil {
		return nil, err
	}
	n := len(submitted)
	if n != 1 {
		return nil, fmt.Errorf("expected one SubmittedExpense, received %d", n)
	}
	return submitted[0], err
}
