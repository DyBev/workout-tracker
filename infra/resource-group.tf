resource "aws_resourcegroups_group" "resource_group" {
  name = local.table_name

  resource_query {
    query = jsonencode({
      ResourceTypeFilters = ["AWS::AllSupported"]
      TagFilters = [
        {
          Key    = "Project"
          Values = [var.project_name]
        },
        {
          Key    = "Environment"
          Values = [var.environment]
        },
        {
          Key    = "Branch"
          Values = [var.branch_name]
        }
      ]
    })
  }
}
