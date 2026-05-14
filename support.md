# gcloud-go Command Support

Global flags: `--project`, `--zone`, `--account`

| Command | Supported flags | Missing flags |
|---|---|---|
| `gcloud auth login` | `--cred-file`, `--brief`, `--update-adc` | `--no-browser`, `--launch-browser`, `--force` |
| `gcloud auth configure-docker REGISTRIES` | `--include-artifact-registry` | (complete) |
| `gcloud auth application-default login` | `--cred-file` | `--client-id-file`, `--scopes`, `--no-browser`, `--launch-browser`, `--disable-quota-project` |
| `gcloud config set PROPERTY VALUE` | `--installation` | (complete) |
| `gcloud info` | `--format`, `--anonymize` | `--show-log`, `--run-diagnostics` |
| `gcloud iam workload-identity-pools create-cred-config PROVIDER` | `--output-file`, `--service-account`, `--credential-source-file`, `--credential-source-url`, `--credential-source-headers`, `--credential-source-type`, `--credential-source-field-name`, `--subject-token-type`, `--executable-command`, `--executable-timeout-millis`, `--executable-output-file`, `--service-account-token-lifetime-seconds`, `--aws` | `--azure`, `--app-id-uri`, `--enable-imdsv2`, `--credential-cert-*` |
| `gcloud compute ssh USER@INSTANCE` | `--tunnel-through-iap`, `--internal-ip`, `--ssh-key-file`, `--command`, `--ssh-flag`, `--dry-run`, `--plain`, `--strict-host-key-checking` | `--ssh-key-expiration` |
| `gcloud compute scp SRC DST` | `--tunnel-through-iap`, `--internal-ip`, `--ssh-key-file`, `--recurse`, `--port`, `--compress`, `--scp-flag` | (complete) |
| `gcloud compute instances start INSTANCE` | `--async` | (complete) |
| `gcloud compute instances stop INSTANCE` | `--async`, `--discard-local-ssd` | `--no-graceful-shutdown` |
| `gcloud compute instances describe INSTANCE` | `--format` | `--view` |
| `gcloud compute instances list` | `--filter`, `--format` | `--view` |
| `gcloud compute instances create INSTANCE` | `--machine-type`, `--network`, `--subnet`, `--image-family`, `--image-project`, `--boot-disk-size`, `--boot-disk-type`, `--tags`, `--metadata`, `--metadata-from-file`, `--service-account`, `--scopes`, `--no-address` | `--accelerator`, `--can-ip-forward`, `--create-disk`, `--custom-cpu`, `--custom-memory`, `--deletion-protection`, `--hostname`, `--image`, `--labels`, `--local-ssd`, `--maintenance-policy`, `--min-cpu-platform`, `--network-tier`, `--preemptible`, `--private-network-ip`, `--shielded-*` |
| `gcloud compute instances delete INSTANCE` | `--quiet`, `--delete-disks`, `--keep-disks` | (complete) |
| `gcloud compute instance-groups unmanaged list-instances GROUP` | | (complete) |
| `gcloud compute instance-groups unmanaged add-instances GROUP` | `--instances` | (complete) |
| `gcloud compute instance-groups managed list-instances GROUP` | `--region`, `--format` | (complete) |
| `gcloud compute instance-groups managed describe GROUP` | `--region` | (complete) |
| `gcloud compute instance-groups managed resize GROUP` | `--region`, `--size` | `--creation-retries` |
| `gcloud compute project-info remove-metadata` | `--keys`, `--all` | (complete) |
| `gcloud compute forwarding-rules list` | `--format` | (complete) |
| `gcloud compute addresses create NAME` | `--region`, `--network-tier`, `--subnet`, `--network`, `--purpose`, `--address-type`, `--addresses`, `--description` | (complete) |
| `gcloud secrets create SECRET_ID` | `--data-file`, `--replication-policy`, `--locations`, `--labels`, `--expire-time`, `--ttl`, `--location` | `--kms-key-name` |
| `gcloud secrets describe SECRET_ID` | `--location` | (complete) |
| `gcloud secrets delete SECRET_ID` | `--quiet`, `--location` | (complete) |
| `gcloud secrets list` | `--filter`, `--format`, `--location` | (complete) |
| `gcloud secrets versions access VERSION` | `--secret`, `--out-file`, `--location` | (complete) |
| `gcloud secrets versions add SECRET_ID` | `--data-file`, `--location` | (complete) |
| `gcloud scheduler jobs describe JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs pause JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs resume JOB_ID` | `--location` | (complete) |
| `gcloud scheduler jobs run JOB_ID` | `--location` | (complete) |
| `gcloud dataflow jobs list` | `--region`, `--format`, `--filter`, `--status`, `--created-after`, `--created-before` | (complete) |
| `gcloud dataflow jobs describe JOB_ID` | `--region`, `--full` | (complete) |
| `gcloud dataflow jobs cancel JOB_ID` | `--region`, `--force` | (complete) |
| `gcloud storage buckets list` | `--format`, `--soft-deleted` | (complete) |
| `gcloud storage cp SOURCE DESTINATION` | `-r`/`--recursive`, `--no-clobber`, `--storage-class` | `--gzip-*`, `--preserve-posix`, `--daisy-chain`, `--manifest-path` |
| `gcloud storage ls [GCS_PATH]` | `-r`/`--recursive`, `--long`, `--json` | `--full`, `--all-versions`, `--buckets`, `--readable-sizes` |
| `gcloud monitoring policies list` | `--format`, `--filter` | (complete) |
| `gcloud monitoring snoozes create` | `--display-name`, `--start-time`, `--end-time`, `--criteria-policies`, `--criteria-filter`, `--snooze-from-file` | (complete) |
| `gcloud monitoring snoozes describe SNOOZE_ID` | | (complete) |
| `gcloud monitoring snoozes list` | `--format` | (complete) |
| `gcloud monitoring snoozes update SNOOZE_ID` | `--display-name`, `--start-time`, `--end-time`, `--snooze-from-file` | (complete) |
| `gcloud monitoring snoozes cancel SNOOZE_ID` | | (complete) |
| `gcloud artifacts docker images scan IMAGE` | `--location`, `--format`, `--remote`, `--async` | `--additional-package-types` |
| `gcloud artifacts docker images list-vulnerabilities SCAN_RESOURCE` | `--format` | (complete) |
| `gcloud redis instances describe INSTANCE_NAME` | `--region` | (complete) |
| `gcloud dataplex datascans run DATASCAN_ID` | `--location` | (complete) |
| `gcloud dataplex datascans jobs list` | `--location`, `--datascan`, `--format` | (complete) |
| `gcloud dataplex datascans jobs describe JOB_ID` | `--location`, `--datascan`, `--view` | (complete) |
