# Import a custom galaxy cluster by its numeric MISP id.
# Bundled clusters (default=true) cannot be imported — the provider will return
# an error if you attempt to import one.
terraform import misp_galaxy_cluster.example 12345
