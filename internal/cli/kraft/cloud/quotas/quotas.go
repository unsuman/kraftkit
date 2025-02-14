// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package quotas

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/cli/kraft/cloud/utils"

	kraftcloud "sdk.kraft.cloud"
)

type QuotasOptions struct {
	Limits   bool   `long:"limits" short:"l" usage:"Show usage limits"`
	Features bool   `long:"features" short:"f" usage:"Show enabled features"`
	Output   string `local:"true" long:"output" short:"o" usage:"Set output format. Options: table,yaml,json,list" default:"list"`

	metro string
	token string
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&QuotasOptions{}, cobra.Command{
		Short:   "View your resource quota on KraftCloud",
		Use:     "quotas",
		Args:    cobra.NoArgs,
		Aliases: []string{"q", "quota"},
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud",
		},
		Long: heredoc.Doc(`
			View your resource quota on KraftCloud.
		`),
		Example: heredoc.Doc(`
			# View your resource quota on KraftCloud
			$ kraft cloud quota

			# View your resource quota on KraftCloud in JSON format
			$ kraft cloud quota -o json
		`),
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *QuotasOptions) Pre(cmd *cobra.Command, _ []string) error {
	err := utils.PopulateMetroToken(cmd, &opts.metro, &opts.token)
	if err != nil {
		return fmt.Errorf("could not populate metro and token: %w", err)
	}

	if !utils.IsValidOutputFormat(opts.Output) {
		return fmt.Errorf("invalid output format: %s", opts.Output)
	}

	if opts.Limits && opts.Features {
		return fmt.Errorf("cannot use both limits and features flags")
	}

	return nil
}

func (opts *QuotasOptions) Run(ctx context.Context, _ []string) error {
	auth, err := config.GetKraftCloudAuthConfig(ctx, opts.token)
	if err != nil {
		return fmt.Errorf("could not retrieve credentials: %w", err)
	}

	client := kraftcloud.NewUsersClient(
		kraftcloud.WithToken(config.GetKraftCloudTokenAuthConfig(*auth)),
	)

	resp, err := client.WithMetro(opts.metro).Quotas(ctx)
	if err != nil {
		return fmt.Errorf("could not get quotas: %w", err)
	}

	if opts.Limits {
		return utils.PrintQuotasLimits(ctx, opts.Output, *resp)
	}

	if opts.Features {
		return utils.PrintQuotasFeatures(ctx, opts.Output, *resp)
	}

	return utils.PrintQuotas(ctx, *auth, opts.Output, *resp)
}
