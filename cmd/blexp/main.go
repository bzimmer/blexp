package main

// Submit expenses via https://integrations.expensify.com/Integration-Server/doc/

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/bzimmer/blexp"
	"github.com/lukasmalkmus/expensify-go"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"
)

func initLogger() {
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)
}

func initConfig(path string) (io.Reader, error) {
	if path == "" {
		usr, err := user.Current()
		if err != nil {
			log.Err(err).Msg("obtaining home directory failed")
			return nil, err
		}
		path = filepath.Join(usr.HomeDir, ".blexp.toml")
	}
	log.Info().Str("path", path).Msg("reading configuration")
	return os.Open(path)
}

func initBlexp(reader io.Reader) (*blexp.Blexp, error) {
	b, err := blexp.New(reader)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func run(c *cli.Context) error {
	reader, err := initConfig(c.Value("config").(string))
	if err != nil {
		return err
	}

	b, err := initBlexp(reader)
	if err != nil {
		return err
	}

	if c.IsSet("list") {
		for key, val := range b.Templates() {
			prefix := ""
			if key == b.Default() {
				prefix = "* "
			}
			fmt.Printf("%s[%-15s] merchant: %s, amount: %d\n", prefix, key, val.Merchant, val.Amount)
		}
		return nil
	}

	if !c.IsSet("force") {
		log.Info().Msg("not submitting expense")
		return nil
	}

	ctx := context.Background()
	var submitted *expensify.SubmittedExpense
	if c.IsSet("expense") {
		submitted, err = b.SubmitExpense(ctx, c.Value("expense").(string))
	} else {
		submitted, err = b.SubmitDefault(ctx)
	}
	if err != nil {
		return err
	}
	log.Info().Str("op", "submitted").Interface("exp", submitted).Send()
	return nil
}

func main() {
	cli.AppHelpTemplate = fmt.Sprintf(`%s
Acquire a UserID and UserSecret at https://integrations.expensify.com/Integration-Server/doc/

Possible expense template values can be found at https://integrations.expensify.com/Integration-Server/doc/#expense-creator
 - All names use SnakeCase

The configuration file format:

	UserID = "some_user_id"
	UserSecret = "lTLF3PjlcdhkcNgwUkPj2Q"
	UserEmail = "me@example"
	Default = "Broadband"

	[Templates]

	[Templates.Broadband]
	Merchant  = "Comcast"
	Amount    = 5000
	Currency  = "USD"
	Category  = "Employee Reimbursement"
	Tag       = "20 - My Department"

	[Templates.Lunch]
	Merchant  = "My Favorite Lunch Place"
	Amount    = 1500
	Currency  = "USD"
	Category  = "Entertainment"
	Tag       = "20 - My Department"

`, cli.AppHelpTemplate)
	app := &cli.App{
		Name:   "blexp",
		Usage:  "submit expenses from the cli",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				DefaultText: "$HOME/.blexp.toml",
				Value:       "",
				Aliases:     []string{"c"},
				Usage:       "Load configuration from `FILE`",
			},
			&cli.BoolFlag{
				Name:    "force",
				Value:   false,
				Aliases: []string{"f"},
				Usage:   "Force expense submission",
			},
			&cli.BoolFlag{
				Name:    "list",
				Value:   false,
				Aliases: []string{"l"},
				Usage:   "List expense templates",
			},
			&cli.StringFlag{
				Name:    "expense",
				Aliases: []string{"e"},
				Usage:   "Submit an expense for the named template",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	os.Exit(0)
}
