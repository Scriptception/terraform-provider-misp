#!/bin/sh
# Import a MISP noticelist by numeric id. After import, set `name` in your
# config to match what MISP reports for that id.
terraform import misp_noticelist.gdpr 1
