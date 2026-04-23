# Look up a galaxy cluster by its numeric id.
# Works for both custom and bundled clusters (read-only access).

data "misp_galaxy_cluster" "example" {
  id = "1"
}

output "cluster_tag_name" {
  value = data.misp_galaxy_cluster.example.tag_name
}
