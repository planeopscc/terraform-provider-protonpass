// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
	"github.com/planeopscc/terraform-provider-protonpass/internal/provider/datasources"
	"github.com/planeopscc/terraform-provider-protonpass/internal/provider/resources"
)

var _ provider.Provider = &ProtonPassProvider{}

// ProtonPassProvider implements the Proton Pass Terraform provider.
type ProtonPassProvider struct {
	version string
}

// ProtonPassProviderModel describes the provider configuration.
type ProtonPassProviderModel struct {
	CLIPath types.String `tfsdk:"cli_path"`
	Timeout types.Int64  `tfsdk:"timeout"`
}

func (p *ProtonPassProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "protonpass"
	resp.Version = p.version
}

func (p *ProtonPassProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Proton Pass. Manages vaults and items via the pass-cli.",
		Attributes: map[string]schema.Attribute{
			"cli_path": schema.StringAttribute{
				MarkdownDescription: "Path to the pass-cli binary. Default: 'pass-cli' (from PATH).",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout in seconds for CLI commands. Default: 30.",
				Optional:            true,
			},
		},
	}
}

func (p *ProtonPassProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProtonPassProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cliPath := "pass-cli"
	if !data.CLIPath.IsNull() && !data.CLIPath.IsUnknown() {
		cliPath = data.CLIPath.ValueString()
	}

	timeout := 30 * time.Second
	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "configuring protonpass provider", map[string]interface{}{
		"pass_cli_path": cliPath,
		"timeout":       timeout.String(),
	})

	runner := passcli.NewExecRunner(cliPath, timeout)
	client := passcli.NewClient(runner)

	if err := client.HealthCheck(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Proton Pass CLI Not Ready",
			fmt.Sprintf("Could not verify pass-cli session. "+
				"Ensure pass-cli is installed at %q and you are logged in (pass-cli login).\n\nError: %s", cliPath, err),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ProtonPassProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewVaultResource,
		resources.NewItemResource,
		resources.NewVaultMemberResource,
		resources.NewAliasResource,
	}
}

func (p *ProtonPassProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewVaultsDataSource,
		datasources.NewItemsDataSource,
		datasources.NewItemDataSource,
		datasources.NewTotpDataSource,
	}
}

// New creates a new provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProtonPassProvider{
			version: version,
		}
	}
}
