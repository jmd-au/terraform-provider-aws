// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearchserverless_lifecycle_policy", name="Lifecycle Policy")
func newLifecyclePolicyDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &lifecyclePolicyDataSource{}, nil
}

const (
	DSNameLifecyclePolicy = "Lifecycle Policy Data Source"
)

type lifecyclePolicyDataSource struct {
	framework.DataSourceWithModel[lifecyclePolicyDataSourceModel]
}

func (d *lifecyclePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedDate: schema.StringAttribute{
				Description: "The date the lifecycle policy was created.",
				Computed:    true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the policy. Typically used to store information about the permissions defined in the policy.",
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_date": schema.StringAttribute{
				Description: "The date the lifecycle policy was last modified.",
				Computed:    true,
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				Description: "JSON policy document to use as the content for the new policy.",
				Computed:    true,
			},
			"policy_version": schema.StringAttribute{
				Description: "Version of the policy.",
				Computed:    true,
			},
			names.AttrType: schema.StringAttribute{
				Description: "Type of lifecycle policy. Must be `retention`.",
				Required:    true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.LifecyclePolicyType](),
				},
			},
		},
	}
}

func (d *lifecyclePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data lifecyclePolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLifecyclePolicyByNameAndType(ctx, conn, data.Name.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameLifecyclePolicy, data.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.Name)
	createdDate := time.UnixMilli(aws.ToInt64(out.CreatedDate))
	data.CreatedDate = flex.StringValueToFramework(ctx, createdDate.Format(time.RFC3339))

	lastModifiedDate := time.UnixMilli(aws.ToInt64(out.LastModifiedDate))
	data.LastModifiedDate = flex.StringValueToFramework(ctx, lastModifiedDate.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type lifecyclePolicyDataSourceModel struct {
	framework.WithRegionModel
	CreatedDate      types.String `tfsdk:"created_date"`
	Description      types.String `tfsdk:"description"`
	ID               types.String `tfsdk:"id"`
	LastModifiedDate types.String `tfsdk:"last_modified_date"`
	Name             types.String `tfsdk:"name"`
	Policy           types.String `tfsdk:"policy"`
	PolicyVersion    types.String `tfsdk:"policy_version"`
	Type             types.String `tfsdk:"type"`
}
