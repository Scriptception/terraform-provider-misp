#!/bin/sh
# Import an existing sharing-group server entry using the composite id:
#   <sharing_group_id>:<server_id>
# Both values are the numeric MISP identifiers (visible in the MISP UI or API).
# Note: server_id="0" is reserved for MISP's local-instance entry and cannot be imported.
terraform import misp_sharing_group_server.peer_in_partners 3:2
