#!/bin/sh
# Import an existing sharing-group membership using the composite id:
#   <sharing_group_id>:<organisation_id>
# Both values are the numeric MISP identifiers (visible in the MISP UI or API).
terraform import misp_sharing_group_member.acme_partner 3:7
