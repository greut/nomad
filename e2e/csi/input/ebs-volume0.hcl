id        = "ebs-vol[0]"
name      = "this-is-a-test-0" # CSIVolumeName tag
type      = "csi"
plugin_id = "aws-ebs0"

capacity_min = "10GiB"
capacity_max = "20GiB"

capability {
  access_mode     = "single-node-writer"
  attachment_mode = "file-system"
}

capability {
  access_mode     = "single-node-writer"
  attachment_mode = "block-device"
}

parameters {
  type = "gp2"
}
