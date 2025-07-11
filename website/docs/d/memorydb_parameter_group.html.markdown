---
subcategory: "MemoryDB"
layout: "aws"
page_title: "AWS: aws_memorydb_parameter_group"
description: |-
  Provides information about a MemoryDB Parameter Group.
---

# Resource: aws_memorydb_parameter_group

Provides information about a MemoryDB Parameter Group.

## Example Usage

```terraform
data "aws_memorydb_parameter_group" "example" {
  name = "my-parameter-group"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the parameter group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the parameter group.
* `arn` - ARN of the parameter group.
* `description` - Description of the parameter group.
* `family` - Engine version that the parameter group can be used with.
* `parameter` - Set of user-defined MemoryDB parameters applied by the parameter group.
    * `name` - Name of the parameter.
    * `value` - Value of the parameter.
* `tags` - Map of tags assigned to the parameter group.
