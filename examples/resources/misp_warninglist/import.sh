#!/bin/sh
# Import a MISP warninglist by numeric id. After import, set `name` in your
# config to match what MISP reports for that id.
terraform import misp_warninglist.akamai 1
