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
	apigateway "google.golang.org/api/apigateway/v1"
	alloydb "google.golang.org/api/alloydb/v1"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
	memcache "google.golang.org/api/memcache/v1"
	recommender "google.golang.org/api/recommender/v1"
	servicedirectory "google.golang.org/api/servicedirectory/v1"
	looker "google.golang.org/api/looker/v1"
	batchapi "google.golang.org/api/batch/v1"
	datalineage "google.golang.org/api/datalineage/v1"
	observability "google.golang.org/api/observability/v1"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	cloudbuild2 "google.golang.org/api/cloudbuild/v2"
	pubsub "google.golang.org/api/pubsub/v1"
	iamv2 "google.golang.org/api/iam/v2"
	accessapproval "google.golang.org/api/accessapproval/v1"
	agentregistry "google.golang.org/api/agentregistry/v1alpha"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	config1 "google.golang.org/api/config/v1"
	publicca "google.golang.org/api/publicca/v1"
	metastore "google.golang.org/api/metastore/v1"
	privateca "google.golang.org/api/privateca/v1"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	networksecurity "google.golang.org/api/networksecurity/v1"
	networksecuritybeta "google.golang.org/api/networksecurity/v1beta1"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
	securitycenter "google.golang.org/api/securitycenter/v1"
	securityposture "google.golang.org/api/securityposture/v1"
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

func BeyondCorpService(ctx context.Context, account string) (*beyondcorp.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return beyondcorp.NewService(ctx, option.WithTokenSource(ts))
}

func CloudBuildService(ctx context.Context, account string) (*cloudbuild.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudbuild.NewService(ctx, option.WithTokenSource(ts))
}

func CloudBuildV2Service(ctx context.Context, account string) (*cloudbuild2.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudbuild2.NewService(ctx, option.WithTokenSource(ts))
}

func PubSubService(ctx context.Context, account string) (*pubsub.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return pubsub.NewService(ctx, option.WithTokenSource(ts))
}

func IAMV2Service(ctx context.Context, account string) (*iamv2.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return iamv2.NewService(ctx, option.WithTokenSource(ts))
}

func AccessApprovalService(ctx context.Context, account string) (*accessapproval.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return accessapproval.NewService(ctx, option.WithTokenSource(ts))
}

func MetastoreService(ctx context.Context, account string) (*metastore.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return metastore.NewService(ctx, option.WithTokenSource(ts))
}

func ObservabilityService(ctx context.Context, account string) (*observability.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return observability.NewService(ctx, option.WithTokenSource(ts))
}

func DataLineageService(ctx context.Context, account string) (*datalineage.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return datalineage.NewService(ctx, option.WithTokenSource(ts))
}

func BatchService(ctx context.Context, account string) (*batchapi.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return batchapi.NewService(ctx, option.WithTokenSource(ts))
}

func LookerService(ctx context.Context, account string) (*looker.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return looker.NewService(ctx, option.WithTokenSource(ts))
}

func ServiceDirectoryService(ctx context.Context, account string) (*servicedirectory.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return servicedirectory.NewService(ctx, option.WithTokenSource(ts))
}

func RecommenderService(ctx context.Context, account string) (*recommender.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return recommender.NewService(ctx, option.WithTokenSource(ts))
}

func MemcacheService(ctx context.Context, account string) (*memcache.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return memcache.NewService(ctx, option.WithTokenSource(ts))
}

func CloudTasksService(ctx context.Context, account string) (*cloudtasks.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudtasks.NewService(ctx, option.WithTokenSource(ts))
}

func AlloyDBService(ctx context.Context, account string) (*alloydb.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return alloydb.NewService(ctx, option.WithTokenSource(ts))
}

func APIGatewayService(ctx context.Context, account string) (*apigateway.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return apigateway.NewService(ctx, option.WithTokenSource(ts))
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

func PublicCAService(ctx context.Context, account string) (*publicca.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return publicca.NewService(ctx, option.WithTokenSource(ts))
}

func InfraManagerService(ctx context.Context, account string) (*config1.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return config1.NewService(ctx, option.WithTokenSource(ts))
}

func AgentRegistryService(ctx context.Context, account string) (*agentregistry.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return agentregistry.NewService(ctx, option.WithTokenSource(ts))
}

func CloudIdentityService(ctx context.Context, account string) (*cloudidentity.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudidentity.NewService(ctx, option.WithTokenSource(ts))
}

func NetworkSecurityService(ctx context.Context, account string) (*networksecurity.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networksecurity.NewService(ctx, option.WithTokenSource(ts))
}

func NetworkSecurityBetaService(ctx context.Context, account string) (*networksecuritybeta.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networksecuritybeta.NewService(ctx, option.WithTokenSource(ts))
}

func SecurityCenterService(ctx context.Context, account string) (*securitycenter.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return securitycenter.NewService(ctx, option.WithTokenSource(ts))
}

func SecurityPostureService(ctx context.Context, account string) (*securityposture.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return securityposture.NewService(ctx, option.WithTokenSource(ts))
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
