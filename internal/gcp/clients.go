package gcp

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	apikeys "google.golang.org/api/apikeys/v2"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	assuredworkloads "google.golang.org/api/assuredworkloads/v1"
	billingbudgets "google.golang.org/api/billingbudgets/v1"
	cloudasset "google.golang.org/api/cloudasset/v1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v3"
	cloudiam "google.golang.org/api/iam/v1"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	dataflow "google.golang.org/api/dataflow/v1b3"
	datamigration "google.golang.org/api/datamigration/v1"
	dataplex "google.golang.org/api/dataplex/v1"
	eventarc "google.golang.org/api/eventarc/v1"
	firestore "google.golang.org/api/firestore/v1"
	aiplatform "google.golang.org/api/aiplatform/v1"
	notebooks "google.golang.org/api/notebooks/v2"
	transcoder "google.golang.org/api/transcoder/v1"
	managedkafka "google.golang.org/api/managedkafka/v1"
	datastream "google.golang.org/api/datastream/v1"
	certificatemanager "google.golang.org/api/certificatemanager/v1"
	apphub "google.golang.org/api/apphub/v1"
	privateca "google.golang.org/api/privateca/v1"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	oslogin "google.golang.org/api/oslogin/v1"
	redis "google.golang.org/api/redis/v1"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
	storage "google.golang.org/api/storage/v1"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

func SchedulerService(ctx context.Context, account string) (*cloudscheduler.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudscheduler.NewService(ctx, option.WithTokenSource(ts))
}

func DataflowService(ctx context.Context, account string) (*dataflow.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return dataflow.NewService(ctx, option.WithTokenSource(ts))
}

func DataMigrationService(ctx context.Context, account string) (*datamigration.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return datamigration.NewService(ctx, option.WithTokenSource(ts))
}

func StorageService(ctx context.Context, account string) (*storage.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return storage.NewService(ctx, option.WithTokenSource(ts))
}

func MonitoringService(ctx context.Context, account string) (*monitoring.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return monitoring.NewService(ctx, option.WithTokenSource(ts))
}

func RedisService(ctx context.Context, account string) (*redis.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return redis.NewService(ctx, option.WithTokenSource(ts))
}

func DataplexService(ctx context.Context, account string) (*dataplex.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return dataplex.NewService(ctx, option.WithTokenSource(ts))
}

func OSLoginService(ctx context.Context, account string) (*oslogin.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return oslogin.NewService(ctx, option.WithTokenSource(ts))
}

func ArtifactRegistryService(ctx context.Context, account string) (*artifactregistry.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return artifactregistry.NewService(ctx, option.WithTokenSource(ts))
}

func IAMService(ctx context.Context, account string) (*cloudiam.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudiam.NewService(ctx, option.WithTokenSource(ts))
}

func CloudResourceManagerService(ctx context.Context, account string) (*cloudresourcemanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudresourcemanager.NewService(ctx, option.WithTokenSource(ts))
}

func AssuredWorkloadsService(ctx context.Context, account string) (*assuredworkloads.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return assuredworkloads.NewService(ctx, option.WithTokenSource(ts))
}

func ServiceUsageService(ctx context.Context, account string) (*serviceusage.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return serviceusage.NewService(ctx, option.WithTokenSource(ts))
}

func APIKeysService(ctx context.Context, account string) (*apikeys.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return apikeys.NewService(ctx, option.WithTokenSource(ts))
}

func ServiceNetworkingService(ctx context.Context, account string) (*servicenetworking.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return servicenetworking.NewService(ctx, option.WithTokenSource(ts))
}

func CloudAssetService(ctx context.Context, account string) (*cloudasset.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudasset.NewService(ctx, option.WithTokenSource(ts))
}

func OrgPolicyService(ctx context.Context, account string) (*orgpolicy.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return orgpolicy.NewService(ctx, option.WithTokenSource(ts))
}

func CloudBillingService(ctx context.Context, account string) (*cloudbilling.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudbilling.NewService(ctx, option.WithTokenSource(ts))
}

func BillingBudgetsService(ctx context.Context, account string) (*billingbudgets.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return billingbudgets.NewService(ctx, option.WithTokenSource(ts))
}

func IAMCredentialsService(ctx context.Context, account string) (*iamcredentials.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return iamcredentials.NewService(ctx, option.WithTokenSource(ts))
}

func OnDemandScanningService(ctx context.Context, account string) (*ondemandscanning.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ondemandscanning.NewService(ctx, option.WithTokenSource(ts))
}

func EventarcService(ctx context.Context, account string) (*eventarc.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return eventarc.NewService(ctx, option.WithTokenSource(ts))
}

func FirestoreService(ctx context.Context, account string) (*firestore.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return firestore.NewService(ctx, option.WithTokenSource(ts))
}

func PolicySimulatorService(ctx context.Context, account string) (*policysimulator.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return policysimulator.NewService(ctx, option.WithTokenSource(ts))
}

func PolicyTroubleshooterService(ctx context.Context, account string) (*policytroubleshooter.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return policytroubleshooter.NewService(ctx, option.WithTokenSource(ts))
}

func PolicyAnalyzerService(ctx context.Context, account string) (*policyanalyzer.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return policyanalyzer.NewService(ctx, option.WithTokenSource(ts))
}

func PrivateCAService(ctx context.Context, account string) (*privateca.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return privateca.NewService(ctx, option.WithTokenSource(ts))
}

func AppHubService(ctx context.Context, account string) (*apphub.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return apphub.NewService(ctx, option.WithTokenSource(ts))
}

func CertificateManagerService(ctx context.Context, account string) (*certificatemanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return certificatemanager.NewService(ctx, option.WithTokenSource(ts))
}

func DatastreamService(ctx context.Context, account string) (*datastream.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return datastream.NewService(ctx, option.WithTokenSource(ts))
}

func ManagedKafkaService(ctx context.Context, account string) (*managedkafka.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return managedkafka.NewService(ctx, option.WithTokenSource(ts))
}

func TranscoderService(ctx context.Context, account string) (*transcoder.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return transcoder.NewService(ctx, option.WithTokenSource(ts))
}

func NotebooksService(ctx context.Context, account string) (*notebooks.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return notebooks.NewService(ctx, option.WithTokenSource(ts))
}

func AIPlatformService(ctx context.Context, account, region string) (*aiplatform.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-aiplatform.googleapis.com/", region)))
	}
	return aiplatform.NewService(ctx, opts...)
}

// PlatformTokenSource returns an OAuth token source with the cloud-platform
// scope for callers that need to make raw HTTP requests against Google Cloud
// endpoints not covered by a generated Go client (for example
// eventarcpublishing.googleapis.com).
func PlatformTokenSource(ctx context.Context, account string) (oauth2.TokenSource, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ts, nil
}
