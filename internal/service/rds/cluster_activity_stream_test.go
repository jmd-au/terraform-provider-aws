// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterActivityStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster types.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterActivityStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterActivityStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterActivityStreamExists(ctx, resourceName, &dbCluster),
					resource.TestCheckResourceAttr(resourceName, "engine_native_audit_fields_included", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "kinesis_stream_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"engine_native_audit_fields_included"},
			},
		},
	})
}

func TestAccRDSClusterActivityStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster types.DBCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_cluster_activity_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterActivityStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterActivityStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterActivityStreamExists(ctx, resourceName, &dbCluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceClusterActivityStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterActivityStreamExists(ctx context.Context, n string, v *types.DBCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBClusterWithActivityStream(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterActivityStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_activity_stream" {
				continue
			}

			_, err := tfrds.FindDBClusterWithActivityStream(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Activity Stream %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccClusterActivityStreamConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
  deletion_protection = false
  engine              = "aurora-postgresql"
  engine_version      = "11.9"
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_rds_cluster.test.id
  engine             = aws_rds_cluster.test.engine
  instance_class     = "db.r6g.large"
}
`, rName))
}

func testAccClusterActivityStreamConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterActivityStreamConfig_base(rName), `
resource "aws_rds_cluster_activity_stream" "test" {
  resource_arn = aws_rds_cluster.test.arn
  kms_key_id   = aws_kms_key.test.key_id
  mode         = "async"

  depends_on = [aws_rds_cluster_instance.test]
}
`)
}
