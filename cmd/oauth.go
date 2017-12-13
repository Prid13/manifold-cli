package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/manifoldco/manifold-cli/analytics"
	"github.com/manifoldco/manifold-cli/api"
	"github.com/manifoldco/manifold-cli/config"
	"github.com/manifoldco/manifold-cli/errs"
	"github.com/manifoldco/manifold-cli/prompts"
	"github.com/manifoldco/manifold-cli/session"
)

var (
	linkedOK = "Your Manifold account is now linked"
)

func init() {
	oauthCmd := cli.Command{
		Name:     "oauth",
		Usage:    "Authenticate with an OAuth provider to register, login or link accounts",
		Category: "AUTHENTICATION",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "github",
				Usage: "Use the web flow for GitHub authentication",
			},
		},
		Action: oauth,
	}

	cmds = append(cmds, oauthCmd)
}

func oauth(cliCtx *cli.Context) error {
	ctx := context.Background()

	if cliCtx.NumFlags() < 1 {
		return errs.NewUsageExitError(cliCtx,
			cli.NewExitError("You must provide an authentication mechanism", -1))
	}

	cfg, a, err := loadConfigAndAnalytics()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Unable to load configuration: %s", err), -1)
	}

	s, err := session.Retrieve(ctx, cfg)
	if err != nil {
		cli.NewExitError(fmt.Sprintf("Could not retrieve session: %s", err), -1)
	}

	if s.Authenticated() {
		// link
		_, err := prompts.Confirm("Do you wish to link your GitHub account to Manifold")
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Could not link accounts: %s", err), -1)
		}

		err = authenticateOAuth(ctx, cliCtx, cfg, a, "link")
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Unable to link accounts: %s", err), -1)
		}

		return cli.NewExitError(linkedOK, 0)
	}

	// registration + login
	err = authenticateOAuth(ctx, cliCtx, cfg, a, "login")
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Could not login with OAuth provider: %s", err), -1)
	}

	fmt.Println("You are logged in, hooray!")
	return nil
}

func loadConfigAndAnalytics() (*config.Config, *analytics.Analytics, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Unable to load config: %s", err), -1)
	}

	a, err := api.New(api.Analytics)
	if err != nil {
		return nil, nil, cli.NewExitError("A problem occurred: "+err.Error(), -1)
	}

	return cfg, a.Analytics, nil
}

func authenticateOAuth(ctx context.Context, cliCtx *cli.Context, cfg *config.Config,
	a *analytics.Analytics, stateType string) error {

	var err error
	if cliCtx.Bool("github") {
		err = githubWithCallback(ctx, cfg, a, stateType)
		if err != nil {
			return err
		}
	}

	return nil
}
