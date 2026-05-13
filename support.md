# gcloud-go Command Support

Global flags: `--project`, `--zone`, `--account`

| Command | Supported flags | Missing flags |
|---|---|---|
| `gcloud auth login` | `--cred-file` | `--no-browser`, `--launch-browser`, `--brief`, `--force`, `--update-adc` |
| `gcloud auth configure-docker REGISTRIES` | | `--include-artifact-registry` |
| `gcloud auth application-default login` | `--cred-file` | `--client-id-file`, `--scopes`, `--no-browser`, `--launch-browser`, `--disable-quota-project` |
| `gcloud config set PROPERTY VALUE` | | `--installation` |
| `gcloud info` | `--format` | `--show-log`, `--run-diagnostics`, `--anonymize` |
| `gcloud iam workload-identity-pools create-cred-config PROVIDER` | `--output-file`, `--service-account`, `--credential-source-file`, `--credential-source-url`, `--credential-source-headers`, `--credential-source-type`, `--credential-source-field-name`, `--subject-token-type`, `--executable-command`, `--executable-timeout-millis`, `--executable-output-file`, `--service-account-token-lifetime-seconds`, `--aws` | `--azure`, `--app-id-uri`, `--enable-imdsv2`, `--credential-cert-*` |
| `gcloud compute ssh USER@INSTANCE` | `--tunnel-through-iap`, `--internal-ip`, `--ssh-key-file` | `--command`, `--ssh-flag`, `--dry-run`, `--plain`, `--strict-host-key-checking`, `--ssh-key-expiration` |
| `gcloud compute scp SRC DST` | `--tunnel-through-iap`, `--internal-ip`, `--ssh-key-file`, `--recurse` | `--port`, `--compress`, `--scp-flag` |
| `gcloud compute instances start INSTANCE` | `--async` | (complete) |
| `gcloud compute instances stop INSTANCE` | `--async` | `--discard-local-ssd`, `--no-graceful-shutdown` |
| `gcloud compute instances describe INSTANCE` | `--format` | `--view` |
| `gcloud compute instances list` | `--filter`, `--format` | `--view` |
| `gcloud compute instances create INSTANCE` | `--machine-type`, `--network`, `--subnet`, `--image-family`, `--image-project`, `--boot-disk-size`, `--boot-disk-type`, `--tags`, `--metadata`, `--metadata-from-file`, `--service-account`, `--scopes`, `--no-address` | `--accelerator`, `--can-ip-forward`, `--create-disk`, `--custom-cpu`, `--custom-memory`, `--deletion-protection`, `--hostname`, `--image`, `--labels`, `--local-ssd`, `--maintenance-policy`, `--min-cpu-platform`, `--network-tier`, `--preemptible`, `--private-network-ip`, `--shielded-*` |
| `gcloud compute instances delete INSTANCE` | `--quiet` | `--delete-disks`, `--keep-disks` |
| `gcloud compute instance-groups unmanaged list-instances GROUP` | | (complete) |
| `gcloud compute instance-groups unmanaged add-instances GROUP` | `--instances` | (complete) |
| `gcloud compute instance-groups managed list-instances GROUP` | `--region`, `--format` | (complete) |
| `gcloud compute instance-groups managed describe GROUP` | `--region` | (complete) |
| `gcloud compute instance-groups managed resize GROUP` | `--region`, `--size` | `--creation-retries` |
| `gcloud compute project-info remove-metadata` | `--keys` | `--all` |
| `gcloud compute forwarding-rules list` | `--format` | (complete) |
| `gcloud compute addresses create NAME` | `--region`, `--network-tier`, `--subnet`, `--network`, `--purpose`, `--address-type` | `--addresses`, `--description` |
| `gcloud secrets create SECRET_ID` | `--data-file` | `--replication-policy`, `--locations`, `--kms-key-name`, `--labels`, `--expire-time`, `--ttl` |
| `gcloud secrets describe SECRET_ID` | | `--location` |
| `gcloud secrets delete SECRET_ID` | `--quiet` | `--location` |
| `gcloud secrets list` | `--filter`, `--format` | `--location` |
| `gcloud secrets versions access VERSION` | `--secret`, `--out-file` | `--location` |
| `gcloud secrets versions add SECRET_ID` | `--data-file` | `--location` |
| `gcloud scheduler jobs describe JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs pause JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs resume JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs run JOB_ID` | `--location` | (complete) |
| `gcloud dataflow jobs list` | `--region`, `--format`, `--filter`, `--status` | `--created-after`, `--created-before` |
| `gcloud dataflow jobs describe JOB_ID` | `--region` | `--full` |
| `gcloud dataflow jobs cancel JOB_ID` | `--region` | `--force` |
| `gcloud storage buckets list` | `--format` | `--soft-deleted` |
| `gcloud storage cp SOURCE DESTINATION` | `-r`/`--recursive` | `--storage-class`, `--gzip-*`, `--preserve-posix`, `--no-clobber`, `--daisy-chain`, `--manifest-path` |
| `gcloud storage ls [GCS_PATH]` | `-r`/`--recursive` | `--long`, `--full`, `--json`, `--all-versions`, `--buckets`, `--readable-sizes` |
| `gcloud monitoring policies list` | `--format`, `--filter` | (complete) |
| `gcloud monitoring snoozes create` | `--display-name`, `--start-time`, `--end-time`, `--criteria-policies`, `--criteria-filter` | `--snooze-from-file` |
| `gcloud monitoring snoozes describe SNOOZE_ID` | | (complete) |
| `gcloud monitoring snoozes list` | `--format` | (complete) |
| `gcloud monitoring snoozes update SNOOZE_ID` | `--display-name`, `--start-time`, `--end-time` | `--snooze-from-file` |
| `gcloud monitoring snoozes cancel SNOOZE_ID` | | (complete) |
| `gcloud artifacts docker images scan IMAGE` | `--location`, `--format` | `--remote`, `--additional-package-types`, `--async` |
| `gcloud artifacts docker images list-vulnerabilities SCAN_RESOURCE` | `--format` | (complete) |
| `gcloud redis instances describe INSTANCE_NAME` | `--region` | (complete) |
| `gcloud dataplex datascans run DATASCAN_ID` | `--location` | (complete) |
| `gcloud dataplex datascans jobs list` | `--location`, `--datascan`, `--format` | (complete) |
| `gcloud dataplex datascans jobs describe JOB_ID` | `--location`, `--datascan` | `--view` |
