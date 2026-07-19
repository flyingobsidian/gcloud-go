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
	cloudlocationfinder "google.golang.org/api/cloudlocationfinder/v1"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	dataflow "google.golang.org/api/dataflow/v1b3"
	datamigration "google.golang.org/api/datamigration/v1"
	dataplex "google.golang.org/api/dataplex/v1"
	datastore "google.golang.org/api/datastore/v1"
	parametermanager "google.golang.org/api/parametermanager/v1"
	eventarc "google.golang.org/api/eventarc/v1"
	firestore "google.golang.org/api/firestore/v1"
	aiplatform "google.golang.org/api/aiplatform/v1"
	aiplatformbeta "google.golang.org/api/aiplatform/v1beta1"
	cloudkms "google.golang.org/api/cloudkms/v1"
	kmsinventory "google.golang.org/api/kmsinventory/v1"
	iap "google.golang.org/api/iap/v1"
	networkservices "google.golang.org/api/networkservices/v1"
	networkservicesbeta "google.golang.org/api/networkservices/v1beta1"
	notebooks "google.golang.org/api/notebooks/v2"
	notebooksv1 "google.golang.org/api/notebooks/v1"
	transcoder "google.golang.org/api/transcoder/v1"
	managedkafka "google.golang.org/api/managedkafka/v1"
	datastream "google.golang.org/api/datastream/v1"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2"
	deploymentmanagerbeta "google.golang.org/api/deploymentmanager/v2beta"
	domains "google.golang.org/api/domains/v1"
	filestore "google.golang.org/api/file/v1"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	networkmanagement "google.golang.org/api/networkmanagement/v1"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
	securesourcemanager "google.golang.org/api/securesourcemanager/v1"
	spanner "google.golang.org/api/spanner/v1"
	certificatemanager "google.golang.org/api/certificatemanager/v1"
	apphub "google.golang.org/api/apphub/v1"
	apigateway "google.golang.org/api/apigateway/v1"
	alloydb "google.golang.org/api/alloydb/v1"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
	memcache "google.golang.org/api/memcache/v1"
	recommender "google.golang.org/api/recommender/v1"
	servicedirectory "google.golang.org/api/servicedirectory/v1"
	looker "google.golang.org/api/looker/v1"
	managedidentities "google.golang.org/api/managedidentities/v1"
	batchapi "google.golang.org/api/batch/v1"
	datalineage "google.golang.org/api/datalineage/v1"
	observability "google.golang.org/api/observability/v1"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	cloudbuild2 "google.golang.org/api/cloudbuild/v2"
	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctionsv2 "google.golang.org/api/cloudfunctions/v2"
	logging "google.golang.org/api/logging/v2"
	pubsub "google.golang.org/api/pubsub/v1"
	iamv2 "google.golang.org/api/iam/v2"
	accessapproval "google.golang.org/api/accessapproval/v1"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
	agentregistry "google.golang.org/api/agentregistry/v1alpha"
	apihub "google.golang.org/api/apihub/v1"
	appengine "google.golang.org/api/appengine/v1"
	ids "google.golang.org/api/ids/v1"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
	composer "google.golang.org/api/composer/v1"
	datacatalog "google.golang.org/api/datacatalog/v1"
	dataproc "google.golang.org/api/dataproc/v1"
	dns "google.golang.org/api/dns/v1"
	netapp "google.golang.org/api/netapp/v1"
	recaptchaenterprise "google.golang.org/api/recaptchaenterprise/v1"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	config1 "google.golang.org/api/config/v1"
	publicca "google.golang.org/api/publicca/v1"
	metastore "google.golang.org/api/metastore/v1"
	ml "google.golang.org/api/ml/v1"
	privateca "google.golang.org/api/privateca/v1"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
	networksecurity "google.golang.org/api/networksecurity/v1"
	networksecuritybeta "google.golang.org/api/networksecurity/v1beta1"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
	securitycenter "google.golang.org/api/securitycenter/v1"
	securityposture "google.golang.org/api/securityposture/v1"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	oslogin "google.golang.org/api/oslogin/v1"
	pubsublite "google.golang.org/api/pubsublite/v1"
	redis "google.golang.org/api/redis/v1"
	runv1 "google.golang.org/api/run/v1"
	runv2 "google.golang.org/api/run/v2"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
	sqladmin "google.golang.org/api/sqladmin/v1"
	storage "google.golang.org/api/storage/v1"
	storagetransfer "google.golang.org/api/storagetransfer/v1"
	vmmigration "google.golang.org/api/vmmigration/v1"
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

func AccessContextManagerService(ctx context.Context, account string) (*accesscontextmanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return accesscontextmanager.NewService(ctx, option.WithTokenSource(ts))
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

// NotebooksV1Service returns a Cloud Notebooks v1 client. The v1 API exposes
// the User-Managed Notebook Environments, Instances, and Managed Runtimes
// surfaces that gcloud's `notebooks` command tree targets; v2 covers only a
// subset (Instances).
func NotebooksV1Service(ctx context.Context, account string) (*notebooksv1.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return notebooksv1.NewService(ctx, option.WithTokenSource(ts))
}

// IAPService returns a Cloud Identity-Aware Proxy v1 client. Backs
// `gcloud iap oauth-brands`, `gcloud iap oauth-clients`, `gcloud iap settings`
// and `gcloud iap tcp dest-groups`.
func IAPService(ctx context.Context, account string) (*iap.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return iap.NewService(ctx, option.WithTokenSource(ts))
}

// NetworkServicesService returns a Network Services v1 client. The v1 client
// covers the IAM-only Media CDN Edge Cache surfaces; the full EdgeCache CRUD
// endpoints (keysets, origins, services) are reached via a raw REST client in
// cmd/edge_cache.go because they are not generated in google.golang.org/api.
func NetworkServicesService(ctx context.Context, account string) (*networkservices.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networkservices.NewService(ctx, option.WithTokenSource(ts))
}

// NetworkServicesBetaService returns a Network Services v1beta1 client. Some
// surfaces (e.g. agent-gateways) are only exposed on v1beta1 as of this build.
func NetworkServicesBetaService(ctx context.Context, account string) (*networkservicesbeta.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networkservicesbeta.NewService(ctx, option.WithTokenSource(ts))
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

// AIPlatformBetaService returns a regional aiplatform v1beta1 client. v1beta1
// exposes surfaces that are not yet in v1 -- most notably
// PublishersModelsService.List for Model Garden.
func AIPlatformBetaService(ctx context.Context, account, region string) (*aiplatformbeta.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-aiplatform.googleapis.com/", region)))
	}
	return aiplatformbeta.NewService(ctx, opts...)
}

// MLService returns a client for the legacy Cloud ML Engine API
// (google.golang.org/api/ml/v1), which backs the `gcloud ai-platform`
// surface. This is distinct from Vertex AI (`gcloud ai ...`), which uses
// the newer aiplatform.googleapis.com API and is region-scoped.
func MLService(ctx context.Context, account string) (*ml.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ml.NewService(ctx, option.WithTokenSource(ts))
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

func NetworkConnectivityService(ctx context.Context, account string) (*networkconnectivity.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networkconnectivity.NewService(ctx, option.WithTokenSource(ts))
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

func DeploymentManagerService(ctx context.Context, account string) (*deploymentmanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return deploymentmanager.NewService(ctx, option.WithTokenSource(ts))
}

func DeploymentManagerBetaService(ctx context.Context, account string) (*deploymentmanagerbeta.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return deploymentmanagerbeta.NewService(ctx, option.WithTokenSource(ts))
}

func CloudFunctionsV1Service(ctx context.Context, account string) (*cloudfunctionsv1.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudfunctionsv1.NewService(ctx, option.WithTokenSource(ts))
}

func CloudFunctionsV2Service(ctx context.Context, account string) (*cloudfunctionsv2.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudfunctionsv2.NewService(ctx, option.WithTokenSource(ts))
}

func AppEngineService(ctx context.Context, account string) (*appengine.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return appengine.NewService(ctx, option.WithTokenSource(ts))
}

func LoggingService(ctx context.Context, account string) (*logging.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return logging.NewService(ctx, option.WithTokenSource(ts))
}

func SQLAdminService(ctx context.Context, account string) (*sqladmin.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return sqladmin.NewService(ctx, option.WithTokenSource(ts))
}

func StorageTransferService(ctx context.Context, account string) (*storagetransfer.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return storagetransfer.NewService(ctx, option.WithTokenSource(ts))
}

func ServiceManagementService(ctx context.Context, account string) (*servicemanagement.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return servicemanagement.NewService(ctx, option.WithTokenSource(ts))
}

func VMMigrationService(ctx context.Context, account string) (*vmmigration.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return vmmigration.NewService(ctx, option.WithTokenSource(ts))
}

func SecureSourceManagerService(ctx context.Context, account string) (*securesourcemanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return securesourcemanager.NewService(ctx, option.WithTokenSource(ts))
}

func DomainsService(ctx context.Context, account string) (*domains.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return domains.NewService(ctx, option.WithTokenSource(ts))
}

func FilestoreService(ctx context.Context, account string) (*filestore.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return filestore.NewService(ctx, option.WithTokenSource(ts))
}

func WorkflowExecutionsService(ctx context.Context, account string) (*workflowexecutions.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return workflowexecutions.NewService(ctx, option.WithTokenSource(ts))
}

func MonitoringV1Service(ctx context.Context, account string) (*monitoringv1.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return monitoringv1.NewService(ctx, option.WithTokenSource(ts))
}

func NetworkManagementService(ctx context.Context, account string) (*networkmanagement.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return networkmanagement.NewService(ctx, option.WithTokenSource(ts))
}

func CloudLocationFinderService(ctx context.Context, account string) (*cloudlocationfinder.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudlocationfinder.NewService(ctx, option.WithTokenSource(ts))
}

func DatastoreService(ctx context.Context, account string) (*datastore.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return datastore.NewService(ctx, option.WithTokenSource(ts))
}

func NetAppService(ctx context.Context, account string) (*netapp.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return netapp.NewService(ctx, option.WithTokenSource(ts))
}

func BigtableAdminService(ctx context.Context, account string) (*bigtableadmin.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return bigtableadmin.NewService(ctx, option.WithTokenSource(ts))
}

func ComposerService(ctx context.Context, account string) (*composer.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return composer.NewService(ctx, option.WithTokenSource(ts))
}

func DataCatalogService(ctx context.Context, account string) (*datacatalog.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return datacatalog.NewService(ctx, option.WithTokenSource(ts))
}

func CloudDeployService(ctx context.Context, account string) (*clouddeploy.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return clouddeploy.NewService(ctx, option.WithTokenSource(ts))
}

func DNSService(ctx context.Context, account string) (*dns.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return dns.NewService(ctx, option.WithTokenSource(ts))
}

func ManagedIdentitiesService(ctx context.Context, account string) (*managedidentities.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return managedidentities.NewService(ctx, option.WithTokenSource(ts))
}

func DataprocService(ctx context.Context, account, region string) (*dataproc.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-dataproc.googleapis.com/", region)))
	}
	return dataproc.NewService(ctx, opts...)
}

func IDSService(ctx context.Context, account string) (*ids.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ids.NewService(ctx, option.WithTokenSource(ts))
}

func ReCaptchaEnterpriseService(ctx context.Context, account string) (*recaptchaenterprise.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return recaptchaenterprise.NewService(ctx, option.WithTokenSource(ts))
}

func ApiHubService(ctx context.Context, account string) (*apihub.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return apihub.NewService(ctx, option.WithTokenSource(ts))
}

// PubSubLiteService returns a pubsublite client wired to the region-specific
// endpoint. Pub/Sub Lite is a regional service: admin operations for a given
// location must be routed through <region>-pubsublite.googleapis.com, where
// the region is derived from the (regional or zonal) location the resource
// lives in.
func PubSubLiteService(ctx context.Context, account, region string) (*pubsublite.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-pubsublite.googleapis.com/", region)))
	}
	return pubsublite.NewService(ctx, opts...)
}

func ParameterManagerService(ctx context.Context, account string) (*parametermanager.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return parametermanager.NewService(ctx, option.WithTokenSource(ts))
}

// RunV1Service returns a Cloud Run v1 (Knative-style) client. Cloud Run's v1
// surface is regional; pass an empty region for the multi-region default
// endpoint or a region name (e.g. "us-central1") to route calls through
// https://REGION-run.googleapis.com/.
func RunV1Service(ctx context.Context, account, region string) (*runv1.APIService, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-run.googleapis.com/", region)))
	}
	return runv1.NewService(ctx, opts...)
}

// RunV2Service returns a Cloud Run v2 client. Cloud Run v2 supports both a
// global and per-region endpoint; pass a region to pin the call to
// https://REGION-run.googleapis.com/ (recommended for regional resources).
func RunV2Service(ctx context.Context, account, region string) (*runv2.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	opts := []option.ClientOption{option.WithTokenSource(ts)}
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-run.googleapis.com/", region)))
	}
	return runv2.NewService(ctx, opts...)
}

// KMSService returns a Cloud KMS v1 client for key rings, keys, versions,
// EKM connections, import jobs, key handles, retired resources, and
// single-tenant HSM management.
func KMSService(ctx context.Context, account string) (*cloudkms.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return cloudkms.NewService(ctx, option.WithTokenSource(ts))
}

// KMSInventoryService returns a KMS Inventory v1 client for listing keys and
// searching resources protected by a Cloud KMS key.
func KMSInventoryService(ctx context.Context, account string) (*kmsinventory.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return kmsinventory.NewService(ctx, option.WithTokenSource(ts))
}

// SpannerService returns a Cloud Spanner v1 client.
func SpannerService(ctx context.Context, account string) (*spanner.Service, error) {
	ts, err := auth.TokenSource(ctx, account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return spanner.NewService(ctx, option.WithTokenSource(ts))
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
