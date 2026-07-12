# Live integration comparison suite

Exercise gcloud-go against real GCP projects and compare output with the reference Python gcloud, over a create -> list -> destroy lifecycle.

The projects are permanent:
- `fo-gcloud-go` - contains non-test long-lived objects, for example the service account that connects Github to GCP.
- `fo-gcloud-py-testci` - for creating, listing and deleting objects using the reference Python gcloud.
- `fo-gcloud-go-testci` - for creating, listing and deleting objects using `gcloud-go`.
