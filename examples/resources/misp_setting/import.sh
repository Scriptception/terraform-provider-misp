#!/bin/sh
# Import a MISP setting by its dotted name. After import, set `value` in
# your configuration to the value you want Terraform to manage.
terraform import misp_setting.baseurl MISP.baseurl
