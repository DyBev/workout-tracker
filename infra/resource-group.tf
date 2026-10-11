resource "aws_resourcegroups_group" "resource_group" {
  name = local.table_name

  resource_query {
    query = <<JSON
{
  "Type": "TAG_FILTERS_1_0",
  "TagFilters": [
    {
      "Key": "Project",
      "Values": ["${var.project_name}"]
    },
    {
      "Key": "Environment",
      "Values": ["${var.environment}"]
    },
    {
      "Key": "Branch",
      "Values": ["${var.branch_name}"]
    }
  ]
}
JSON
  }
}
