// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_resourceexplorer2_search", name="Search")
func newSearchDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &searchDataSource{}, nil
}

const (
	DSNameSearch = "Search Data Source"
)

type searchDataSource struct {
	framework.DataSourceWithModel[searchDataSourceModel]
}

func (d *searchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"query_string": schema.StringAttribute{
				Required: true,
			},
			"resource_count":    framework.DataSourceComputedListOfObjectAttribute[countData](ctx),
			names.AttrResources: framework.DataSourceComputedListOfObjectAttribute[resourcesData](ctx),
			"view_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1011),
				},
			},
		},
	}
}

func (d *searchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ResourceExplorer2Client(ctx)

	var data searchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ViewArn.IsNull() {
		data.ID = types.StringValue(fmt.Sprintf(",%s", data.QueryString.ValueString()))
	} else {
		data.ID = types.StringValue(fmt.Sprintf("%s,%s", data.ViewArn.ValueString(), data.QueryString.ValueString()))
	}

	input := &resourceexplorer2.SearchInput{
		QueryString: data.QueryString.ValueStringPointer(),
	}
	if !data.ViewArn.IsNull() {
		input.ViewArn = data.ViewArn.ValueStringPointer()
	}

	paginator := resourceexplorer2.NewSearchPaginator(conn, input)

	var out resourceexplorer2.SearchOutput
	commonFieldsSet := false
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResourceExplorer2, create.ErrActionReading, DSNameSearch, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.Resources) > 0 {
			if !commonFieldsSet {
				out.Count = page.Count
				out.ViewArn = page.ViewArn
				commonFieldsSet = true
			}
			out.Resources = append(out.Resources, page.Resources...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type searchDataSourceModel struct {
	framework.WithRegionModel
	Count       fwtypes.ListNestedObjectValueOf[countData]     `tfsdk:"resource_count"`
	ID          types.String                                   `tfsdk:"id"`
	QueryString types.String                                   `tfsdk:"query_string"`
	Resources   fwtypes.ListNestedObjectValueOf[resourcesData] `tfsdk:"resources"`
	ViewArn     fwtypes.ARN                                    `tfsdk:"view_arn"`
}

type countData struct {
	Complete       types.Bool  `tfsdk:"complete"`
	TotalResources types.Int64 `tfsdk:"total_resources"`
}

type resourcesData struct {
	ARN             fwtypes.ARN                                     `tfsdk:"arn"`
	LastReportedAt  timetypes.RFC3339                               `tfsdk:"last_reported_at"`
	OwningAccountID types.String                                    `tfsdk:"owning_account_id"`
	Properties      fwtypes.ListNestedObjectValueOf[propertiesData] `tfsdk:"properties"`
	Region          types.String                                    `tfsdk:"region"`
	ResourceType    types.String                                    `tfsdk:"resource_type"`
	Service         types.String                                    `tfsdk:"service"`
}

type propertiesData struct {
	Data           jsontypes.Normalized `tfsdk:"data"`
	LastReportedAt timetypes.RFC3339    `tfsdk:"last_reported_at"`
	Name           types.String         `tfsdk:"name"`
}
