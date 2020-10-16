package main

// Submit expenses via https://integrations.expensify.com/Integration-Server/doc/

import (
	"context"
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"github.com/bzimmer/blexp"
	"github.com/lukasmalkmus/expensify-go"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"
)

var b *blexp.Blexp

type config struct {
	UserID     string                        `json:"user_id"`
	UserSecret string                        `json:"user_secret"`
	UserEmail  string                        `json:"user_email"`
	Primary    string                        `json:"primary"`
	Templates  *map[string]expensify.Expense `json:"templates"`
}

func defaultConfig() string {
	usr, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msg("obtaining home directory failed")
		return ""
	}
	return filepath.Join(usr.HomeDir, ".blexp.json")
}

func readConfig(path string) (*config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	c := &config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func initLogging(c *cli.Context) error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)
	log.Debug().Msg("configured logging")
	return nil
}

func initBlexp(c *cli.Context) error {
	path := c.Value("config").(string)
	log.Info().Str("path", path).Msg("reading configuration")
	cfg, err := readConfig(path)
	if err != nil {
		return err
	}
	b, err = blexp.New(cfg.UserID, cfg.UserSecret,
		blexp.WithTemplates(*cfg.Templates, cfg.Primary),
		blexp.WithUserEmail(cfg.UserEmail))
	if err != nil {
		return err
	}
	return nil
}

func list(c *cli.Context) error {
	for key, val := range b.Templates {
		log.Info().
			Str("name", key).
			Bool("primary", key == b.Primary).
			Interface("template", val).
			Msg("list")
	}
	return nil
}

func submit(c *cli.Context) error {
	s := c.Args().Slice()
	if len(s) == 0 {
		s = make([]string, 1)
		s[0] = b.Primary
	}

	force := c.IsSet("force")
	ctx := context.Background()
	for _, x := range s {
		entry := log.Warn()
		exp, err := b.PrepareExpense(x)
		var submitted *expensify.SubmittedExpense
		if exp != nil {
			if force {
				entry = log.Info()
				submitted, err = b.SubmitExpense(ctx, exp)
			} else {
				submitted, err = &expensify.SubmittedExpense{Expense: *exp}, nil
			}
		}
		entry.
			Err(err).
			Str("template", x).
			Bool("submitted", force).
			Interface("exp", submitted).
			Msg("submit")
	}
	return nil
}

func main() {
	var configFlag = &cli.StringFlag{
		Name:    "config",
		Hidden:  false,
		Value:   defaultConfig(),
		Aliases: []string{"c"},
		Usage:   "Load configuration from `FILE`",
	}
	app := &cli.App{
		Name:   "blexp",
		Usage:  "submit expenses from the cli",
		Before: initLogging,
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List expense templates",
				Before:  initBlexp,
				Action:  list,
				Flags: []cli.Flag{
					configFlag,
				},
			},
			{
				Name:    "submit",
				Aliases: []string{"s"},
				Usage:   "Submit expenses",
				Before:  initBlexp,
				Action:  submit,
				Flags: []cli.Flag{
					configFlag,
					&cli.BoolFlag{
						Name:    "force",
						Value:   false,
						Aliases: []string{"f"},
						Usage:   "Force expense submission",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	os.Exit(0)
}
