package gcp

import (
	"context"
	"io"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"golang.org/x/oauth2"
	accessapproval "google.golang.org/api/accessapproval/v1"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
	agentidentity "google.golang.org/api/agentidentity/v1"
	agentidentityalpha "google.golang.org/api/agentidentity/v1alpha"
	agentregistry "google.golang.org/api/agentregistry/v1alpha"
	aiplatform "google.golang.org/api/aiplatform/v1"
	aiplatformbeta "google.golang.org/api/aiplatform/v1beta1"
	alloydb "google.golang.org/api/alloydb/v1"
	apigateway "google.golang.org/api/apigateway/v1"
	apigee "google.golang.org/api/apigee/v1"
	apihub "google.golang.org/api/apihub/v1"
	apikeys "google.golang.org/api/apikeys/v2"
	appengine "google.golang.org/api/appengine/v1"
	apphub "google.golang.org/api/apphub/v1"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	assuredworkloads "google.golang.org/api/assuredworkloads/v1"
	auditmanager "google.golang.org/api/auditmanager/v1"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
	batchapi "google.golang.org/api/batch/v1"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
	billingbudgets "google.golang.org/api/billingbudgets/v1"
	binaryauthorization "google.golang.org/api/binaryauthorization/v1"
	certificatemanager "google.golang.org/api/certificatemanager/v1"
	cloudasset "google.golang.org/api/cloudasset/v1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	cloudbuild "google.golang.org/api/cloudbuild/v1"
	cloudbuild2 "google.golang.org/api/cloudbuild/v2"
	clouddeploy "google.golang.org/api/clouddeploy/v1"
	cloudfunctionsv1 "google.golang.org/api/cloudfunctions/v1"
	cloudfunctionsv2 "google.golang.org/api/cloudfunctions/v2"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	cloudkms "google.golang.org/api/cloudkms/v1"
	cloudlocationfinder "google.golang.org/api/cloudlocationfinder/v1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v3"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
	composer "google.golang.org/api/composer/v1"
	composerbeta "google.golang.org/api/composer/v1beta1"
	config1 "google.golang.org/api/config/v1"
	container "google.golang.org/api/container/v1"
	datacatalog "google.golang.org/api/datacatalog/v1"
	dataflow "google.golang.org/api/dataflow/v1b3"
	datalineage "google.golang.org/api/datalineage/v1"
	datamigration "google.golang.org/api/datamigration/v1"
	dataplex "google.golang.org/api/dataplex/v1"
	dataproc "google.golang.org/api/dataproc/v1"
	datastore "google.golang.org/api/datastore/v1"
	datastream "google.golang.org/api/datastream/v1"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2"
	deploymentmanagerbeta "google.golang.org/api/deploymentmanager/v2beta"
	developerconnect "google.golang.org/api/developerconnect/v1"
	dns "google.golang.org/api/dns/v1"
	domains "google.golang.org/api/domains/v1"
	eventarc "google.golang.org/api/eventarc/v1"
	filestore "google.golang.org/api/file/v1"
	firestore "google.golang.org/api/firestore/v1"
	gkehub "google.golang.org/api/gkehub/v1"
	gkeonprem "google.golang.org/api/gkeonprem/v1"
	healthcare "google.golang.org/api/healthcare/v1"
	cloudiam "google.golang.org/api/iam/v1"
	iamv2 "google.golang.org/api/iam/v2"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
	iap "google.golang.org/api/iap/v1"
	ids "google.golang.org/api/ids/v1"
	kmsinventory "google.golang.org/api/kmsinventory/v1"
	logging "google.golang.org/api/logging/v2"
	looker "google.golang.org/api/looker/v1"
	managedidentities "google.golang.org/api/managedidentities/v1"
	managedkafka "google.golang.org/api/managedkafka/v1"
	memcache "google.golang.org/api/memcache/v1"
	metastore "google.golang.org/api/metastore/v1"
	ml "google.golang.org/api/ml/v1"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	netapp "google.golang.org/api/netapp/v1"
	networkconnectivity "google.golang.org/api/networkconnectivity/v1"
	networkmanagement "google.golang.org/api/networkmanagement/v1"
	networksecurity "google.golang.org/api/networksecurity/v1"
	networksecuritybeta "google.golang.org/api/networksecurity/v1beta1"
	networkservices "google.golang.org/api/networkservices/v1"
	networkservicesbeta "google.golang.org/api/networkservices/v1beta1"
	notebooksv1 "google.golang.org/api/notebooks/v1"
	notebooks "google.golang.org/api/notebooks/v2"
	observability "google.golang.org/api/observability/v1"
	ondemandscanning "google.golang.org/api/ondemandscanning/v1"
	oracledatabase "google.golang.org/api/oracledatabase/v1"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	oslogin "google.golang.org/api/oslogin/v1"
	parametermanager "google.golang.org/api/parametermanager/v1"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
	privateca "google.golang.org/api/privateca/v1"
	publicca "google.golang.org/api/publicca/v1"
	pubsub "google.golang.org/api/pubsub/v1"
	pubsublite "google.golang.org/api/pubsublite/v1"
	recaptchaenterprise "google.golang.org/api/recaptchaenterprise/v1"
	recommender "google.golang.org/api/recommender/v1"
	redis "google.golang.org/api/redis/v1"
	runv1 "google.golang.org/api/run/v1"
	runv2 "google.golang.org/api/run/v2"
	securesourcemanager "google.golang.org/api/securesourcemanager/v1"
	securitycenter "google.golang.org/api/securitycenter/v1"
	securityposture "google.golang.org/api/securityposture/v1"
	servicedirectory "google.golang.org/api/servicedirectory/v1"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
	servicenetworking "google.golang.org/api/servicenetworking/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
	sourcerepo "google.golang.org/api/sourcerepo/v1"
	spanner "google.golang.org/api/spanner/v1"
	sqladmin "google.golang.org/api/sqladmin/v1"
	storage "google.golang.org/api/storage/v1"
	storagebatchoperations "google.golang.org/api/storagebatchoperations/v1"
	storagetransfer "google.golang.org/api/storagetransfer/v1"
	testing "google.golang.org/api/testing/v1"
	transcoder "google.golang.org/api/transcoder/v1"
	vmmigration "google.golang.org/api/vmmigration/v1"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
	workflows "google.golang.org/api/workflows/v1"
	workstations "google.golang.org/api/workstations/v1"
)

// The functions in this file build clients from the process environment with
// a fresh, uncached TokenProvider, exactly as before Clients existed. They
// keep unmigrated command groups working (#2070) and will be removed once
// every caller holds a Clients.

func legacyClients(account string) *Clients {
	return &Clients{Tokens: auth.NewTokenProvider(config.OSEnv()), Account: account}
}

// ApigeeServiceForOrg is the legacy form of Clients.ApigeeForOrg.
func ApigeeServiceForOrg(ctx context.Context, account, org string) (*apigee.Service, error) {
	return legacyClients(account).ApigeeForOrg(ctx, org)
}

// ApigeeOrgHost is the legacy form of Clients.ApigeeOrgHost.
func ApigeeOrgHost(ctx context.Context, account, org string) (string, error) {
	return legacyClients(account).ApigeeOrgHost(ctx, org)
}

// SecurityCenterService is the legacy form of Clients.SecurityCenter. When
// debugOut is non-nil each outgoing request is written to it as a curl
// command (see internal/httplog) before being sent.
func SecurityCenterService(ctx context.Context, account string, debugOut io.Writer) (*securitycenter.Service, error) {
	c := legacyClients(account)
	c.DebugOut = debugOut
	return c.SecurityCenter(ctx)
}

// PlatformTokenSource is the legacy form of Clients.PlatformTokenSource.
func PlatformTokenSource(ctx context.Context, account string) (oauth2.TokenSource, error) {
	return legacyClients(account).PlatformTokenSource(ctx)
}

// SchedulerService is the legacy form of Clients.Scheduler.
func SchedulerService(ctx context.Context, account string) (*cloudscheduler.Service, error) {
	return legacyClients(account).Scheduler(ctx)
}

// DataflowService is the legacy form of Clients.Dataflow.
func DataflowService(ctx context.Context, account string) (*dataflow.Service, error) {
	return legacyClients(account).Dataflow(ctx)
}

// DataMigrationService is the legacy form of Clients.DataMigration.
func DataMigrationService(ctx context.Context, account string) (*datamigration.Service, error) {
	return legacyClients(account).DataMigration(ctx)
}

// StorageService is the legacy form of Clients.Storage.
func StorageService(ctx context.Context, account string) (*storage.Service, error) {
	return legacyClients(account).Storage(ctx)
}

// MonitoringService is the legacy form of Clients.Monitoring.
func MonitoringService(ctx context.Context, account string) (*monitoring.Service, error) {
	return legacyClients(account).Monitoring(ctx)
}

// RedisService is the legacy form of Clients.Redis.
func RedisService(ctx context.Context, account string) (*redis.Service, error) {
	return legacyClients(account).Redis(ctx)
}

// DataplexService is the legacy form of Clients.Dataplex.
func DataplexService(ctx context.Context, account string) (*dataplex.Service, error) {
	return legacyClients(account).Dataplex(ctx)
}

// OSLoginService is the legacy form of Clients.OSLogin.
func OSLoginService(ctx context.Context, account string) (*oslogin.Service, error) {
	return legacyClients(account).OSLogin(ctx)
}

// ArtifactRegistryService is the legacy form of Clients.ArtifactRegistry.
func ArtifactRegistryService(ctx context.Context, account string) (*artifactregistry.Service, error) {
	return legacyClients(account).ArtifactRegistry(ctx)
}

// IAMService is the legacy form of Clients.IAM.
func IAMService(ctx context.Context, account string) (*cloudiam.Service, error) {
	return legacyClients(account).IAM(ctx)
}

// CloudResourceManagerService is the legacy form of Clients.CloudResourceManager.
func CloudResourceManagerService(ctx context.Context, account string) (*cloudresourcemanager.Service, error) {
	return legacyClients(account).CloudResourceManager(ctx)
}

// AuditManagerService is the legacy form of Clients.AuditManager.
func AuditManagerService(ctx context.Context, account string) (*auditmanager.Service, error) {
	return legacyClients(account).AuditManager(ctx)
}

// ServiceUsageService is the legacy form of Clients.ServiceUsage.
func ServiceUsageService(ctx context.Context, account string) (*serviceusage.Service, error) {
	return legacyClients(account).ServiceUsage(ctx)
}

// APIKeysService is the legacy form of Clients.APIKeys.
func APIKeysService(ctx context.Context, account string) (*apikeys.Service, error) {
	return legacyClients(account).APIKeys(ctx)
}

// ServiceNetworkingService is the legacy form of Clients.ServiceNetworking.
func ServiceNetworkingService(ctx context.Context, account string) (*servicenetworking.APIService, error) {
	return legacyClients(account).ServiceNetworking(ctx)
}

// CloudAssetService is the legacy form of Clients.CloudAsset.
func CloudAssetService(ctx context.Context, account string) (*cloudasset.Service, error) {
	return legacyClients(account).CloudAsset(ctx)
}

// OrgPolicyService is the legacy form of Clients.OrgPolicy.
func OrgPolicyService(ctx context.Context, account string) (*orgpolicy.Service, error) {
	return legacyClients(account).OrgPolicy(ctx)
}

// CloudBillingService is the legacy form of Clients.CloudBilling.
func CloudBillingService(ctx context.Context, account string) (*cloudbilling.APIService, error) {
	return legacyClients(account).CloudBilling(ctx)
}

// BillingBudgetsService is the legacy form of Clients.BillingBudgets.
func BillingBudgetsService(ctx context.Context, account string) (*billingbudgets.Service, error) {
	return legacyClients(account).BillingBudgets(ctx)
}

// IAMCredentialsService is the legacy form of Clients.IAMCredentials.
func IAMCredentialsService(ctx context.Context, account string) (*iamcredentials.Service, error) {
	return legacyClients(account).IAMCredentials(ctx)
}

// OnDemandScanningService is the legacy form of Clients.OnDemandScanning.
func OnDemandScanningService(ctx context.Context, account string) (*ondemandscanning.Service, error) {
	return legacyClients(account).OnDemandScanning(ctx)
}

// EventarcService is the legacy form of Clients.Eventarc.
func EventarcService(ctx context.Context, account string) (*eventarc.Service, error) {
	return legacyClients(account).Eventarc(ctx)
}

// FirestoreService is the legacy form of Clients.Firestore.
func FirestoreService(ctx context.Context, account string) (*firestore.Service, error) {
	return legacyClients(account).Firestore(ctx)
}

// PolicySimulatorService is the legacy form of Clients.PolicySimulator.
func PolicySimulatorService(ctx context.Context, account string) (*policysimulator.Service, error) {
	return legacyClients(account).PolicySimulator(ctx)
}

// PolicyTroubleshooterService is the legacy form of Clients.PolicyTroubleshooter.
func PolicyTroubleshooterService(ctx context.Context, account string) (*policytroubleshooter.Service, error) {
	return legacyClients(account).PolicyTroubleshooter(ctx)
}

// PolicyAnalyzerService is the legacy form of Clients.PolicyAnalyzer.
func PolicyAnalyzerService(ctx context.Context, account string) (*policyanalyzer.Service, error) {
	return legacyClients(account).PolicyAnalyzer(ctx)
}

// PrivateCAService is the legacy form of Clients.PrivateCA.
func PrivateCAService(ctx context.Context, account string) (*privateca.Service, error) {
	return legacyClients(account).PrivateCA(ctx)
}

// BeyondCorpService is the legacy form of Clients.BeyondCorp.
func BeyondCorpService(ctx context.Context, account string) (*beyondcorp.Service, error) {
	return legacyClients(account).BeyondCorp(ctx)
}

// CloudBuildService is the legacy form of Clients.CloudBuild.
func CloudBuildService(ctx context.Context, account string) (*cloudbuild.Service, error) {
	return legacyClients(account).CloudBuild(ctx)
}

// CloudBuildV2Service is the legacy form of Clients.CloudBuildV2.
func CloudBuildV2Service(ctx context.Context, account string) (*cloudbuild2.Service, error) {
	return legacyClients(account).CloudBuildV2(ctx)
}

// PubSubService is the legacy form of Clients.PubSub.
func PubSubService(ctx context.Context, account string) (*pubsub.Service, error) {
	return legacyClients(account).PubSub(ctx)
}

// IAMV2Service is the legacy form of Clients.IAMV2.
func IAMV2Service(ctx context.Context, account string) (*iamv2.Service, error) {
	return legacyClients(account).IAMV2(ctx)
}

// AccessApprovalService is the legacy form of Clients.AccessApproval.
func AccessApprovalService(ctx context.Context, account string) (*accessapproval.Service, error) {
	return legacyClients(account).AccessApproval(ctx)
}

// AccessContextManagerService is the legacy form of Clients.AccessContextManager.
func AccessContextManagerService(ctx context.Context, account string) (*accesscontextmanager.Service, error) {
	return legacyClients(account).AccessContextManager(ctx)
}

// MetastoreService is the legacy form of Clients.Metastore.
func MetastoreService(ctx context.Context, account string) (*metastore.APIService, error) {
	return legacyClients(account).Metastore(ctx)
}

// ObservabilityService is the legacy form of Clients.Observability.
func ObservabilityService(ctx context.Context, account string) (*observability.Service, error) {
	return legacyClients(account).Observability(ctx)
}

// DataLineageService is the legacy form of Clients.DataLineage.
func DataLineageService(ctx context.Context, account string) (*datalineage.Service, error) {
	return legacyClients(account).DataLineage(ctx)
}

// BatchService is the legacy form of Clients.Batch.
func BatchService(ctx context.Context, account string) (*batchapi.Service, error) {
	return legacyClients(account).Batch(ctx)
}

// LookerService is the legacy form of Clients.Looker.
func LookerService(ctx context.Context, account string) (*looker.Service, error) {
	return legacyClients(account).Looker(ctx)
}

// ServiceDirectoryService is the legacy form of Clients.ServiceDirectory.
func ServiceDirectoryService(ctx context.Context, account string) (*servicedirectory.APIService, error) {
	return legacyClients(account).ServiceDirectory(ctx)
}

// RecommenderService is the legacy form of Clients.Recommender.
func RecommenderService(ctx context.Context, account string) (*recommender.Service, error) {
	return legacyClients(account).Recommender(ctx)
}

// MemcacheService is the legacy form of Clients.Memcache.
func MemcacheService(ctx context.Context, account string) (*memcache.Service, error) {
	return legacyClients(account).Memcache(ctx)
}

// CloudTasksService is the legacy form of Clients.CloudTasks.
func CloudTasksService(ctx context.Context, account string) (*cloudtasks.Service, error) {
	return legacyClients(account).CloudTasks(ctx)
}

// AlloyDBService is the legacy form of Clients.AlloyDB.
func AlloyDBService(ctx context.Context, account string) (*alloydb.Service, error) {
	return legacyClients(account).AlloyDB(ctx)
}

// APIGatewayService is the legacy form of Clients.APIGateway.
func APIGatewayService(ctx context.Context, account string) (*apigateway.Service, error) {
	return legacyClients(account).APIGateway(ctx)
}

// AppHubService is the legacy form of Clients.AppHub.
func AppHubService(ctx context.Context, account string) (*apphub.APIService, error) {
	return legacyClients(account).AppHub(ctx)
}

// CertificateManagerService is the legacy form of Clients.CertificateManager.
func CertificateManagerService(ctx context.Context, account string) (*certificatemanager.Service, error) {
	return legacyClients(account).CertificateManager(ctx)
}

// DatastreamService is the legacy form of Clients.Datastream.
func DatastreamService(ctx context.Context, account string) (*datastream.Service, error) {
	return legacyClients(account).Datastream(ctx)
}

// ManagedKafkaService is the legacy form of Clients.ManagedKafka.
func ManagedKafkaService(ctx context.Context, account string) (*managedkafka.Service, error) {
	return legacyClients(account).ManagedKafka(ctx)
}

// TranscoderService is the legacy form of Clients.Transcoder.
func TranscoderService(ctx context.Context, account string) (*transcoder.Service, error) {
	return legacyClients(account).Transcoder(ctx)
}

// NotebooksService is the legacy form of Clients.Notebooks.
func NotebooksService(ctx context.Context, account string) (*notebooks.Service, error) {
	return legacyClients(account).Notebooks(ctx)
}

// NotebooksV1Service is the legacy form of Clients.NotebooksV1.
func NotebooksV1Service(ctx context.Context, account string) (*notebooksv1.Service, error) {
	return legacyClients(account).NotebooksV1(ctx)
}

// IAPService is the legacy form of Clients.IAP.
func IAPService(ctx context.Context, account string) (*iap.Service, error) {
	return legacyClients(account).IAP(ctx)
}

// NetworkServicesService is the legacy form of Clients.NetworkServices.
func NetworkServicesService(ctx context.Context, account string) (*networkservices.Service, error) {
	return legacyClients(account).NetworkServices(ctx)
}

// NetworkServicesBetaService is the legacy form of Clients.NetworkServicesBeta.
func NetworkServicesBetaService(ctx context.Context, account string) (*networkservicesbeta.Service, error) {
	return legacyClients(account).NetworkServicesBeta(ctx)
}

// DeveloperConnectService is the legacy form of Clients.DeveloperConnect.
func DeveloperConnectService(ctx context.Context, account string) (*developerconnect.Service, error) {
	return legacyClients(account).DeveloperConnect(ctx)
}

// WorkstationsService is the legacy form of Clients.Workstations.
func WorkstationsService(ctx context.Context, account string) (*workstations.Service, error) {
	return legacyClients(account).Workstations(ctx)
}

// SourceRepoService is the legacy form of Clients.SourceRepo.
func SourceRepoService(ctx context.Context, account string) (*sourcerepo.Service, error) {
	return legacyClients(account).SourceRepo(ctx)
}

// BareMetalSolutionService is the legacy form of Clients.BareMetalSolution.
func BareMetalSolutionService(ctx context.Context, account string) (*baremetalsolution.Service, error) {
	return legacyClients(account).BareMetalSolution(ctx)
}

// HealthcareService is the legacy form of Clients.Healthcare.
func HealthcareService(ctx context.Context, account string) (*healthcare.Service, error) {
	return legacyClients(account).Healthcare(ctx)
}

// VmwareEngineService is the legacy form of Clients.VmwareEngine.
func VmwareEngineService(ctx context.Context, account string) (*vmwareengine.Service, error) {
	return legacyClients(account).VmwareEngine(ctx)
}

// ApigeeService is the legacy form of Clients.Apigee.
func ApigeeService(ctx context.Context, account string) (*apigee.Service, error) {
	return legacyClients(account).Apigee(ctx)
}

// StorageBatchOperationsService is the legacy form of Clients.StorageBatchOperations.
func StorageBatchOperationsService(ctx context.Context, account string) (*storagebatchoperations.Service, error) {
	return legacyClients(account).StorageBatchOperations(ctx)
}

// OracleDatabaseService is the legacy form of Clients.OracleDatabase.
func OracleDatabaseService(ctx context.Context, account string) (*oracledatabase.Service, error) {
	return legacyClients(account).OracleDatabase(ctx)
}

// ContainerService is the legacy form of Clients.Container.
func ContainerService(ctx context.Context, account string) (*container.Service, error) {
	return legacyClients(account).Container(ctx)
}

// GKEHubService is the legacy form of Clients.GKEHub.
func GKEHubService(ctx context.Context, account string) (*gkehub.Service, error) {
	return legacyClients(account).GKEHub(ctx)
}

// GKEOnPremService is the legacy form of Clients.GKEOnPrem.
func GKEOnPremService(ctx context.Context, account string) (*gkeonprem.Service, error) {
	return legacyClients(account).GKEOnPrem(ctx)
}

// BinaryAuthorizationService is the legacy form of Clients.BinaryAuthorization.
func BinaryAuthorizationService(ctx context.Context, account string) (*binaryauthorization.Service, error) {
	return legacyClients(account).BinaryAuthorization(ctx)
}

// TestingService is the legacy form of Clients.Testing.
func TestingService(ctx context.Context, account string) (*testing.APIService, error) {
	return legacyClients(account).Testing(ctx)
}

// MLService is the legacy form of Clients.ML.
func MLService(ctx context.Context, account string) (*ml.Service, error) {
	return legacyClients(account).ML(ctx)
}

// PublicCAService is the legacy form of Clients.PublicCA.
func PublicCAService(ctx context.Context, account string) (*publicca.Service, error) {
	return legacyClients(account).PublicCA(ctx)
}

// InfraManagerService is the legacy form of Clients.InfraManager.
func InfraManagerService(ctx context.Context, account string) (*config1.Service, error) {
	return legacyClients(account).InfraManager(ctx)
}

// AgentRegistryService is the legacy form of Clients.AgentRegistry.
func AgentRegistryService(ctx context.Context, account string) (*agentregistry.APIService, error) {
	return legacyClients(account).AgentRegistry(ctx)
}

// AgentIdentityService is the legacy form of Clients.AgentIdentity.
func AgentIdentityService(ctx context.Context, account string) (*agentidentity.Service, error) {
	return legacyClients(account).AgentIdentity(ctx)
}

// CloudIdentityService is the legacy form of Clients.CloudIdentity.
func CloudIdentityService(ctx context.Context, account string) (*cloudidentity.Service, error) {
	return legacyClients(account).CloudIdentity(ctx)
}

// NetworkConnectivityService is the legacy form of Clients.NetworkConnectivity.
func NetworkConnectivityService(ctx context.Context, account string) (*networkconnectivity.Service, error) {
	return legacyClients(account).NetworkConnectivity(ctx)
}

// NetworkSecurityService is the legacy form of Clients.NetworkSecurity.
func NetworkSecurityService(ctx context.Context, account string) (*networksecurity.Service, error) {
	return legacyClients(account).NetworkSecurity(ctx)
}

// NetworkSecurityBetaService is the legacy form of Clients.NetworkSecurityBeta.
func NetworkSecurityBetaService(ctx context.Context, account string) (*networksecuritybeta.Service, error) {
	return legacyClients(account).NetworkSecurityBeta(ctx)
}

// SecurityPostureService is the legacy form of Clients.SecurityPosture.
func SecurityPostureService(ctx context.Context, account string) (*securityposture.Service, error) {
	return legacyClients(account).SecurityPosture(ctx)
}

// DeploymentManagerService is the legacy form of Clients.DeploymentManager.
func DeploymentManagerService(ctx context.Context, account string) (*deploymentmanager.Service, error) {
	return legacyClients(account).DeploymentManager(ctx)
}

// DeploymentManagerBetaService is the legacy form of Clients.DeploymentManagerBeta.
func DeploymentManagerBetaService(ctx context.Context, account string) (*deploymentmanagerbeta.Service, error) {
	return legacyClients(account).DeploymentManagerBeta(ctx)
}

// CloudFunctionsV1Service is the legacy form of Clients.CloudFunctionsV1.
func CloudFunctionsV1Service(ctx context.Context, account string) (*cloudfunctionsv1.Service, error) {
	return legacyClients(account).CloudFunctionsV1(ctx)
}

// CloudFunctionsV2Service is the legacy form of Clients.CloudFunctionsV2.
func CloudFunctionsV2Service(ctx context.Context, account string) (*cloudfunctionsv2.Service, error) {
	return legacyClients(account).CloudFunctionsV2(ctx)
}

// AppEngineService is the legacy form of Clients.AppEngine.
func AppEngineService(ctx context.Context, account string) (*appengine.APIService, error) {
	return legacyClients(account).AppEngine(ctx)
}

// LoggingService is the legacy form of Clients.Logging.
func LoggingService(ctx context.Context, account string) (*logging.Service, error) {
	return legacyClients(account).Logging(ctx)
}

// SQLAdminService is the legacy form of Clients.SQLAdmin.
func SQLAdminService(ctx context.Context, account string) (*sqladmin.Service, error) {
	return legacyClients(account).SQLAdmin(ctx)
}

// StorageTransferService is the legacy form of Clients.StorageTransfer.
func StorageTransferService(ctx context.Context, account string) (*storagetransfer.Service, error) {
	return legacyClients(account).StorageTransfer(ctx)
}

// ServiceManagementService is the legacy form of Clients.ServiceManagement.
func ServiceManagementService(ctx context.Context, account string) (*servicemanagement.APIService, error) {
	return legacyClients(account).ServiceManagement(ctx)
}

// VMMigrationService is the legacy form of Clients.VMMigration.
func VMMigrationService(ctx context.Context, account string) (*vmmigration.Service, error) {
	return legacyClients(account).VMMigration(ctx)
}

// SecureSourceManagerService is the legacy form of Clients.SecureSourceManager.
func SecureSourceManagerService(ctx context.Context, account string) (*securesourcemanager.Service, error) {
	return legacyClients(account).SecureSourceManager(ctx)
}

// DomainsService is the legacy form of Clients.Domains.
func DomainsService(ctx context.Context, account string) (*domains.Service, error) {
	return legacyClients(account).Domains(ctx)
}

// FilestoreService is the legacy form of Clients.Filestore.
func FilestoreService(ctx context.Context, account string) (*filestore.Service, error) {
	return legacyClients(account).Filestore(ctx)
}

// WorkflowExecutionsService is the legacy form of Clients.WorkflowExecutions.
func WorkflowExecutionsService(ctx context.Context, account string) (*workflowexecutions.Service, error) {
	return legacyClients(account).WorkflowExecutions(ctx)
}

// WorkflowsService is the legacy form of Clients.Workflows.
func WorkflowsService(ctx context.Context, account string) (*workflows.Service, error) {
	return legacyClients(account).Workflows(ctx)
}

// MonitoringV1Service is the legacy form of Clients.MonitoringV1.
func MonitoringV1Service(ctx context.Context, account string) (*monitoringv1.Service, error) {
	return legacyClients(account).MonitoringV1(ctx)
}

// NetworkManagementService is the legacy form of Clients.NetworkManagement.
func NetworkManagementService(ctx context.Context, account string) (*networkmanagement.Service, error) {
	return legacyClients(account).NetworkManagement(ctx)
}

// CloudLocationFinderService is the legacy form of Clients.CloudLocationFinder.
func CloudLocationFinderService(ctx context.Context, account string) (*cloudlocationfinder.Service, error) {
	return legacyClients(account).CloudLocationFinder(ctx)
}

// DatastoreService is the legacy form of Clients.Datastore.
func DatastoreService(ctx context.Context, account string) (*datastore.Service, error) {
	return legacyClients(account).Datastore(ctx)
}

// NetAppService is the legacy form of Clients.NetApp.
func NetAppService(ctx context.Context, account string) (*netapp.Service, error) {
	return legacyClients(account).NetApp(ctx)
}

// BigtableAdminService is the legacy form of Clients.BigtableAdmin.
func BigtableAdminService(ctx context.Context, account string) (*bigtableadmin.Service, error) {
	return legacyClients(account).BigtableAdmin(ctx)
}

// ComposerService is the legacy form of Clients.Composer.
func ComposerService(ctx context.Context, account string) (*composer.Service, error) {
	return legacyClients(account).Composer(ctx)
}

// ComposerBetaService is the legacy form of Clients.ComposerBeta.
func ComposerBetaService(ctx context.Context, account string) (*composerbeta.Service, error) {
	return legacyClients(account).ComposerBeta(ctx)
}

// DataCatalogService is the legacy form of Clients.DataCatalog.
func DataCatalogService(ctx context.Context, account string) (*datacatalog.Service, error) {
	return legacyClients(account).DataCatalog(ctx)
}

// CloudDeployService is the legacy form of Clients.CloudDeploy.
func CloudDeployService(ctx context.Context, account string) (*clouddeploy.Service, error) {
	return legacyClients(account).CloudDeploy(ctx)
}

// DNSService is the legacy form of Clients.DNS.
func DNSService(ctx context.Context, account string) (*dns.Service, error) {
	return legacyClients(account).DNS(ctx)
}

// ManagedIdentitiesService is the legacy form of Clients.ManagedIdentities.
func ManagedIdentitiesService(ctx context.Context, account string) (*managedidentities.Service, error) {
	return legacyClients(account).ManagedIdentities(ctx)
}

// IDSService is the legacy form of Clients.IDS.
func IDSService(ctx context.Context, account string) (*ids.Service, error) {
	return legacyClients(account).IDS(ctx)
}

// ReCaptchaEnterpriseService is the legacy form of Clients.ReCaptchaEnterprise.
func ReCaptchaEnterpriseService(ctx context.Context, account string) (*recaptchaenterprise.Service, error) {
	return legacyClients(account).ReCaptchaEnterprise(ctx)
}

// ApiHubService is the legacy form of Clients.ApiHub.
func ApiHubService(ctx context.Context, account string) (*apihub.Service, error) {
	return legacyClients(account).ApiHub(ctx)
}

// ParameterManagerService is the legacy form of Clients.ParameterManager.
func ParameterManagerService(ctx context.Context, account string) (*parametermanager.Service, error) {
	return legacyClients(account).ParameterManager(ctx)
}

// KMSService is the legacy form of Clients.KMS.
func KMSService(ctx context.Context, account string) (*cloudkms.Service, error) {
	return legacyClients(account).KMS(ctx)
}

// KMSInventoryService is the legacy form of Clients.KMSInventory.
func KMSInventoryService(ctx context.Context, account string) (*kmsinventory.Service, error) {
	return legacyClients(account).KMSInventory(ctx)
}

// SpannerService is the legacy form of Clients.Spanner.
func SpannerService(ctx context.Context, account string) (*spanner.Service, error) {
	return legacyClients(account).Spanner(ctx)
}

// AssuredWorkloadsService is the legacy form of Clients.AssuredWorkloads.
func AssuredWorkloadsService(ctx context.Context, account, endpoint string) (*assuredworkloads.Service, error) {
	return legacyClients(account).AssuredWorkloads(ctx, endpoint)
}

// AIPlatformService is the legacy form of Clients.AIPlatform.
func AIPlatformService(ctx context.Context, account, region string) (*aiplatform.Service, error) {
	return legacyClients(account).AIPlatform(ctx, region)
}

// AIPlatformBetaService is the legacy form of Clients.AIPlatformBeta.
func AIPlatformBetaService(ctx context.Context, account, region string) (*aiplatformbeta.Service, error) {
	return legacyClients(account).AIPlatformBeta(ctx, region)
}

// AgentIdentityAlphaService is the legacy form of Clients.AgentIdentityAlpha.
func AgentIdentityAlphaService(ctx context.Context, account, endpoint string) (*agentidentityalpha.Service, error) {
	return legacyClients(account).AgentIdentityAlpha(ctx, endpoint)
}

// DNSServiceAt is the legacy form of Clients.DNSAt.
func DNSServiceAt(ctx context.Context, account, endpoint string) (*dns.Service, error) {
	return legacyClients(account).DNSAt(ctx, endpoint)
}

// DataprocService is the legacy form of Clients.Dataproc.
func DataprocService(ctx context.Context, account, region string) (*dataproc.Service, error) {
	return legacyClients(account).Dataproc(ctx, region)
}

// PubSubLiteService is the legacy form of Clients.PubSubLite.
func PubSubLiteService(ctx context.Context, account, region string) (*pubsublite.Service, error) {
	return legacyClients(account).PubSubLite(ctx, region)
}

// RunV1Service is the legacy form of Clients.RunV1.
func RunV1Service(ctx context.Context, account, region string) (*runv1.APIService, error) {
	return legacyClients(account).RunV1(ctx, region)
}

// RunV2Service is the legacy form of Clients.RunV2.
func RunV2Service(ctx context.Context, account, region string) (*runv2.Service, error) {
	return legacyClients(account).RunV2(ctx, region)
}
