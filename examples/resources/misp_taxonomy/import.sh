#!/bin/sh
# Import a MISP taxonomy by numeric id. After import, set `namespace` in your
# config to match what MISP reports for that id.
terraform import misp_taxonomy.tlp 5
