// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_location_geofence_collection", name="Geofence Collection")
func DataSourceGeofenceCollection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGeofenceCollectionRead,

		Schema: map[string]*schema.Schema{
			"collection_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"collection_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameGeofenceCollection = "Geofence Collection Data Source"
)

func dataSourceGeofenceCollectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	name := d.Get("collection_name").(string)

	out, err := findGeofenceCollectionByName(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, DSNameGeofenceCollection, name, err)
	}

	d.SetId(aws.ToString(out.CollectionName))
	d.Set("collection_arn", out.CollectionArn)
	d.Set(names.AttrCreateTime, aws.ToTime(out.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrKMSKeyID, out.KmsKeyId)
	d.Set("update_time", aws.ToTime(out.UpdateTime).Format(time.RFC3339))

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	if err := d.Set(names.AttrTags, keyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionSetting, DSNameGeofenceCollection, d.Id(), err)
	}

	return diags
}
