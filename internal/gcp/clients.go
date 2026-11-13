package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/httplog"
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
	compute "google.golang.org/api/compute/v1"
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
	"google.golang.org/api/googleapi"
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
	"google.golang.org/api/option"
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
	secretmanager "google.golang.org/api/secretmanager/v1"
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
	htransport "google.golang.org/api/transport/http"
	vmmigration "google.golang.org/api/vmmigration/v1"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
	workflowexecutions "google.golang.org/api/workflowexecutions/v1"
	workflows "google.golang.org/api/workflows/v1"
	workstations "google.golang.org/api/workstations/v1"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// Scheduler returns a client for the Scheduler API.
func (c *Clients) Scheduler(ctx context.Context) (*cloudscheduler.Service, error) {
	return newService(ctx, c, cloudscheduler.NewService)
}

// Dataflow returns a client for the Dataflow API.
func (c *Clients) Dataflow(ctx context.Context) (*dataflow.Service, error) {
	return newService(ctx, c, dataflow.NewService)
}

// DataMigration returns a client for the DataMigration API.
func (c *Clients) DataMigration(ctx context.Context) (*datamigration.Service, error) {
	return newService(ctx, c, datamigration.NewService)
}

// Storage returns a client for the Storage API.
func (c *Clients) Storage(ctx context.Context) (*storage.Service, error) {
	return newService(ctx, c, storage.NewService)
}

// Monitoring returns a client for the Monitoring API.
func (c *Clients) Monitoring(ctx context.Context) (*monitoring.Service, error) {
	return newService(ctx, c, monitoring.NewService)
}

// Redis returns a client for the Redis API.
func (c *Clients) Redis(ctx context.Context) (*redis.Service, error) {
	return newService(ctx, c, redis.NewService)
}

// Dataplex returns a client for the Dataplex API.
func (c *Clients) Dataplex(ctx context.Context) (*dataplex.Service, error) {
	return newService(ctx, c, dataplex.NewService)
}

// Compute returns a client for the Compute Engine API.
func (c *Clients) Compute(ctx context.Context) (*compute.Service, error) {
	return newService(ctx, c, compute.NewService)
}

// SecretManager returns a client for the Secret Manager API.
func (c *Clients) SecretManager(ctx context.Context) (*secretmanager.Service, error) {
	return newService(ctx, c, secretmanager.NewService)
}

// OSLogin returns a client for the OSLogin API.
func (c *Clients) OSLogin(ctx context.Context) (*oslogin.Service, error) {
	return newService(ctx, c, oslogin.NewService)
}

// ArtifactRegistry returns a client for the ArtifactRegistry API.
func (c *Clients) ArtifactRegistry(ctx context.Context) (*artifactregistry.Service, error) {
	return newService(ctx, c, artifactregistry.NewService)
}

// IAM returns a client for the IAM API.
func (c *Clients) IAM(ctx context.Context) (*cloudiam.Service, error) {
	return newService(ctx, c, cloudiam.NewService)
}

// CloudResourceManager returns a client for the CloudResourceManager API.
func (c *Clients) CloudResourceManager(ctx context.Context) (*cloudresourcemanager.Service, error) {
	return newService(ctx, c, cloudresourcemanager.NewService)
}

// AuditManager returns an Audit Manager v1 client.
func (c *Clients) AuditManager(ctx context.Context) (*auditmanager.Service, error) {
	return newService(ctx, c, auditmanager.NewService)
}

// AssuredWorkloads returns an Assured Workloads v1 client. A non-empty
// endpoint (e.g. https://REGION-assuredworkloads.googleapis.com/) overrides
// the global default.
func (c *Clients) AssuredWorkloads(ctx context.Context, endpoint string) (*assuredworkloads.Service, error) {
	var opts []option.ClientOption
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	return newService(ctx, c, assuredworkloads.NewService, opts...)
}

// ServiceUsage returns a client for the ServiceUsage API.
func (c *Clients) ServiceUsage(ctx context.Context) (*serviceusage.Service, error) {
	return newService(ctx, c, serviceusage.NewService)
}

// APIKeys returns a client for the APIKeys API.
func (c *Clients) APIKeys(ctx context.Context) (*apikeys.Service, error) {
	return newService(ctx, c, apikeys.NewService)
}

// ServiceNetworking returns a client for the ServiceNetworking API.
func (c *Clients) ServiceNetworking(ctx context.Context) (*servicenetworking.APIService, error) {
	return newService(ctx, c, servicenetworking.NewService)
}

// CloudAsset returns a client for the CloudAsset API.
func (c *Clients) CloudAsset(ctx context.Context) (*cloudasset.Service, error) {
	return newService(ctx, c, cloudasset.NewService)
}

// OrgPolicy returns a client for the OrgPolicy API.
func (c *Clients) OrgPolicy(ctx context.Context) (*orgpolicy.Service, error) {
	return newService(ctx, c, orgpolicy.NewService)
}

// CloudBilling returns a client for the CloudBilling API.
func (c *Clients) CloudBilling(ctx context.Context) (*cloudbilling.APIService, error) {
	return newService(ctx, c, cloudbilling.NewService)
}

// BillingBudgets returns a client for the BillingBudgets API.
func (c *Clients) BillingBudgets(ctx context.Context) (*billingbudgets.Service, error) {
	return newService(ctx, c, billingbudgets.NewService)
}

// IAMCredentials returns a client for the IAMCredentials API.
func (c *Clients) IAMCredentials(ctx context.Context) (*iamcredentials.Service, error) {
	return newService(ctx, c, iamcredentials.NewService)
}

// OnDemandScanning returns a client for the OnDemandScanning API.
func (c *Clients) OnDemandScanning(ctx context.Context) (*ondemandscanning.Service, error) {
	return newService(ctx, c, ondemandscanning.NewService)
}

// Eventarc returns a client for the Eventarc API.
func (c *Clients) Eventarc(ctx context.Context) (*eventarc.Service, error) {
	return newService(ctx, c, eventarc.NewService)
}

// Firestore returns a client for the Firestore API.
func (c *Clients) Firestore(ctx context.Context) (*firestore.Service, error) {
	return newService(ctx, c, firestore.NewService)
}

// PolicySimulator returns a client for the PolicySimulator API.
func (c *Clients) PolicySimulator(ctx context.Context) (*policysimulator.Service, error) {
	return newService(ctx, c, policysimulator.NewService)
}

// PolicyTroubleshooter returns a client for the PolicyTroubleshooter API.
func (c *Clients) PolicyTroubleshooter(ctx context.Context) (*policytroubleshooter.Service, error) {
	return newService(ctx, c, policytroubleshooter.NewService)
}

// PolicyAnalyzer returns a client for the PolicyAnalyzer API.
func (c *Clients) PolicyAnalyzer(ctx context.Context) (*policyanalyzer.Service, error) {
	return newService(ctx, c, policyanalyzer.NewService)
}

// PrivateCA returns a client for the PrivateCA API.
func (c *Clients) PrivateCA(ctx context.Context) (*privateca.Service, error) {
	return newService(ctx, c, privateca.NewService)
}

// BeyondCorp returns a client for the BeyondCorp API.
func (c *Clients) BeyondCorp(ctx context.Context) (*beyondcorp.Service, error) {
	return newService(ctx, c, beyondcorp.NewService)
}

// CloudBuild returns a client for the CloudBuild API.
func (c *Clients) CloudBuild(ctx context.Context) (*cloudbuild.Service, error) {
	return newService(ctx, c, cloudbuild.NewService)
}

// CloudBuildV2 returns a client for the CloudBuildV2 API.
func (c *Clients) CloudBuildV2(ctx context.Context) (*cloudbuild2.Service, error) {
	return newService(ctx, c, cloudbuild2.NewService)
}

// PubSub returns a client for the PubSub API.
func (c *Clients) PubSub(ctx context.Context) (*pubsub.Service, error) {
	return newService(ctx, c, pubsub.NewService)
}

// IAMV2 returns a client for the IAMV2 API.
func (c *Clients) IAMV2(ctx context.Context) (*iamv2.Service, error) {
	return newService(ctx, c, iamv2.NewService)
}

// AccessApproval returns a client for the AccessApproval API.
func (c *Clients) AccessApproval(ctx context.Context) (*accessapproval.Service, error) {
	return newService(ctx, c, accessapproval.NewService)
}

// AccessContextManager returns a client for the AccessContextManager API.
func (c *Clients) AccessContextManager(ctx context.Context) (*accesscontextmanager.Service, error) {
	return newService(ctx, c, accesscontextmanager.NewService)
}

// Metastore returns a client for the Metastore API.
func (c *Clients) Metastore(ctx context.Context) (*metastore.APIService, error) {
	return newService(ctx, c, metastore.NewService)
}

// Observability returns a client for the Observability API.
func (c *Clients) Observability(ctx context.Context) (*observability.Service, error) {
	return newService(ctx, c, observability.NewService)
}

// DataLineage returns a client for the DataLineage API.
func (c *Clients) DataLineage(ctx context.Context) (*datalineage.Service, error) {
	return newService(ctx, c, datalineage.NewService)
}

// Batch returns a client for the Batch API.
func (c *Clients) Batch(ctx context.Context) (*batchapi.Service, error) {
	return newService(ctx, c, batchapi.NewService)
}

// Looker returns a client for the Looker API.
func (c *Clients) Looker(ctx context.Context) (*looker.Service, error) {
	return newService(ctx, c, looker.NewService)
}

// ServiceDirectory returns a client for the ServiceDirectory API.
func (c *Clients) ServiceDirectory(ctx context.Context) (*servicedirectory.APIService, error) {
	return newService(ctx, c, servicedirectory.NewService)
}

// Recommender returns a client for the Recommender API.
func (c *Clients) Recommender(ctx context.Context) (*recommender.Service, error) {
	return newService(ctx, c, recommender.NewService)
}

// Memcache returns a client for the Memcache API.
func (c *Clients) Memcache(ctx context.Context) (*memcache.Service, error) {
	return newService(ctx, c, memcache.NewService)
}

// CloudTasks returns a client for the CloudTasks API.
func (c *Clients) CloudTasks(ctx context.Context) (*cloudtasks.Service, error) {
	return newService(ctx, c, cloudtasks.NewService)
}

// AlloyDB returns a client for the AlloyDB API.
func (c *Clients) AlloyDB(ctx context.Context) (*alloydb.Service, error) {
	return newService(ctx, c, alloydb.NewService)
}

// APIGateway returns a client for the APIGateway API.
func (c *Clients) APIGateway(ctx context.Context) (*apigateway.Service, error) {
	return newService(ctx, c, apigateway.NewService)
}

// AppHub returns a client for the AppHub API.
func (c *Clients) AppHub(ctx context.Context) (*apphub.APIService, error) {
	return newService(ctx, c, apphub.NewService)
}

// CertificateManager returns a client for the CertificateManager API.
func (c *Clients) CertificateManager(ctx context.Context) (*certificatemanager.Service, error) {
	return newService(ctx, c, certificatemanager.NewService)
}

// Datastream returns a client for the Datastream API.
func (c *Clients) Datastream(ctx context.Context) (*datastream.Service, error) {
	return newService(ctx, c, datastream.NewService)
}

// ManagedKafka returns a client for the ManagedKafka API.
func (c *Clients) ManagedKafka(ctx context.Context) (*managedkafka.Service, error) {
	return newService(ctx, c, managedkafka.NewService)
}

// Transcoder returns a client for the Transcoder API.
func (c *Clients) Transcoder(ctx context.Context) (*transcoder.Service, error) {
	return newService(ctx, c, transcoder.NewService)
}

// Notebooks returns a client for the Notebooks API.
func (c *Clients) Notebooks(ctx context.Context) (*notebooks.Service, error) {
	return newService(ctx, c, notebooks.NewService)
}

// NotebooksV1 returns a Cloud Notebooks v1 client. The v1 API exposes
// the User-Managed Notebook Environments, Instances, and Managed Runtimes
// surfaces that gcloud's `notebooks` command tree targets; v2 covers only a
// subset (Instances).
func (c *Clients) NotebooksV1(ctx context.Context) (*notebooksv1.Service, error) {
	return newService(ctx, c, notebooksv1.NewService)
}

// IAP returns a Cloud Identity-Aware Proxy v1 client. Backs
// `gcloud iap oauth-brands`, `gcloud iap oauth-clients`, `gcloud iap settings`
// and `gcloud iap tcp dest-groups`.
func (c *Clients) IAP(ctx context.Context) (*iap.Service, error) {
	return newService(ctx, c, iap.NewService)
}

// NetworkServices returns a Network Services v1 client. The v1 client
// covers the IAM-only Media CDN Edge Cache surfaces; the full EdgeCache CRUD
// endpoints (keysets, origins, services) are reached via a raw REST client in
// cmd/edge_cache.go because they are not generated in google.golang.org/api.
func (c *Clients) NetworkServices(ctx context.Context) (*networkservices.Service, error) {
	return newService(ctx, c, networkservices.NewService)
}

// NetworkServicesBeta returns a Network Services v1beta1 client. Some
// surfaces (e.g. agent-gateways) are only exposed on v1beta1 as of this build.
func (c *Clients) NetworkServicesBeta(ctx context.Context) (*networkservicesbeta.Service, error) {
	return newService(ctx, c, networkservicesbeta.NewService)
}

// DeveloperConnect returns a Developer Connect v1 client.
func (c *Clients) DeveloperConnect(ctx context.Context) (*developerconnect.Service, error) {
	return newService(ctx, c, developerconnect.NewService)
}

// Workstations returns a Cloud Workstations v1 client.
func (c *Clients) Workstations(ctx context.Context) (*workstations.Service, error) {
	return newService(ctx, c, workstations.NewService)
}

// SourceRepo returns a Cloud Source Repositories v1 client.
func (c *Clients) SourceRepo(ctx context.Context) (*sourcerepo.Service, error) {
	return newService(ctx, c, sourcerepo.NewService)
}

// BareMetalSolution returns a Bare Metal Solution v2 client.
func (c *Clients) BareMetalSolution(ctx context.Context) (*baremetalsolution.Service, error) {
	return newService(ctx, c, baremetalsolution.NewService)
}

// Healthcare returns a Cloud Healthcare API v1 client.
func (c *Clients) Healthcare(ctx context.Context) (*healthcare.Service, error) {
	return newService(ctx, c, healthcare.NewService)
}

// VmwareEngine returns a VMware Engine v1 client.
func (c *Clients) VmwareEngine(ctx context.Context) (*vmwareengine.Service, error) {
	return newService(ctx, c, vmwareengine.NewService)
}

// Apigee returns an Apigee v1 client.
func (c *Clients) Apigee(ctx context.Context) (*apigee.Service, error) {
	return newService(ctx, c, apigee.NewService)
}

// ApigeeForOrg returns an Apigee v1 client routed to the endpoint that
// serves the given organization's location (see ApigeeOrgHost). org may be a
// bare ID or "organizations/ORG".
func (c *Clients) ApigeeForOrg(ctx context.Context, org string) (*apigee.Service, error) {
	host, err := c.ApigeeOrgHost(ctx, org)
	if err != nil {
		return nil, err
	}
	return newService(ctx, c, apigee.NewService, option.WithEndpoint("https://"+host+"/"))
}

// ApigeeOrgHost returns the Apigee API host serving the given organization,
// mirroring gcloud-python's `project_mapping.ResolveLocation` +
// `request._GetApigeeHostByLocation`: the organization's location is looked
// up via `organizations.getProjectMapping` on the global endpoint and mapped
// to a host by ApigeeHostForLocation.
func (c *Clients) ApigeeOrgHost(ctx context.Context, org string) (string, error) {
	global, err := c.Apigee(ctx)
	if err != nil {
		return "", err
	}
	org = strings.TrimPrefix(org, "organizations/")
	denied := fmt.Errorf("Permission denied on resource \"organizations/%s\" (or it may not exist)", org)
	mapping, err := global.Organizations.GetProjectMapping("organizations/" + org).Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && (gerr.Code == http.StatusNotFound ||
			gerr.Code == http.StatusForbidden || gerr.Code == http.StatusUnauthorized) {
			return "", denied
		}
		return "", fmt.Errorf("getting project mapping: %w", err)
	}
	if mapping.Organization == "" {
		return "", denied
	}
	return ApigeeHostForLocation(mapping.Location), nil
}

const apigeeGlobalHost = "apigee.googleapis.com"

// apigeeMREPRegions mirrors SUPPORTED_MREP_REGIONS in gcloud-python's
// command_lib/apigee/request.py ("ch" added in 586.0.0 for the Switzerland
// data residency zone).
var apigeeMREPRegions = map[string]bool{"us": true, "eu": true, "in": true, "sa": true, "ch": true}

// ApigeeHostForLocation returns the Apigee API host for an organization
// location: the global host for "" or "global", the multi-region REP host
// apigee.LOCATION.rep.googleapis.com for MREP regions, and the locational
// host LOCATION-apigee.googleapis.com otherwise.
func ApigeeHostForLocation(location string) string {
	switch {
	case location == "" || location == "global":
		return apigeeGlobalHost
	case apigeeMREPRegions[location]:
		return fmt.Sprintf("apigee.%s.rep.googleapis.com", location)
	default:
		return fmt.Sprintf("%s-apigee.googleapis.com", location)
	}
}

// StorageBatchOperations returns a Storage Batch Operations v1 client.
func (c *Clients) StorageBatchOperations(ctx context.Context) (*storagebatchoperations.Service, error) {
	return newService(ctx, c, storagebatchoperations.NewService)
}

// OracleDatabase returns an Oracle Database v1 client.
func (c *Clients) OracleDatabase(ctx context.Context) (*oracledatabase.Service, error) {
	return newService(ctx, c, oracledatabase.NewService)
}

// Container returns a GKE (container) v1 client.
func (c *Clients) Container(ctx context.Context) (*container.Service, error) {
	return newService(ctx, c, container.NewService)
}

// GKEHub returns a GKE Hub (fleet) v1 client.
func (c *Clients) GKEHub(ctx context.Context) (*gkehub.Service, error) {
	return newService(ctx, c, gkehub.NewService)
}

// GKEOnPrem returns an Anthos-on-Prem (gkeonprem) v1 client. It
// backs `container bare-metal` and `container vmware`.
func (c *Clients) GKEOnPrem(ctx context.Context) (*gkeonprem.Service, error) {
	return newService(ctx, c, gkeonprem.NewService)
}

// BinaryAuthorization returns a Binary Authorization v1 client.
func (c *Clients) BinaryAuthorization(ctx context.Context) (*binaryauthorization.Service, error) {
	return newService(ctx, c, binaryauthorization.NewService)
}

// Testing returns a Firebase Test Lab v1 client.
func (c *Clients) Testing(ctx context.Context) (*testing.APIService, error) {
	return newService(ctx, c, testing.NewService)
}

func (c *Clients) AIPlatform(ctx context.Context, region string) (*aiplatform.Service, error) {
	var opts []option.ClientOption
	if endpoint := aiplatformEndpointForRegion(region); endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	return newService(ctx, c, aiplatform.NewService, opts...)
}

// AIPlatformBeta returns a regional aiplatform v1beta1 client. v1beta1
// exposes surfaces that are not yet in v1 -- most notably
// PublishersModelsService.List for Model Garden.
func (c *Clients) AIPlatformBeta(ctx context.Context, region string) (*aiplatformbeta.Service, error) {
	var opts []option.ClientOption
	if endpoint := aiplatformEndpointForRegion(region); endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	return newService(ctx, c, aiplatformbeta.NewService, opts...)
}

// aiplatformEndpointForRegion returns the aiplatform endpoint override for a
// given region. A GCP region (e.g. "us-central1") is served on the locational
// host <region>-aiplatform.googleapis.com. Multi-regional (mREP) jurisdictions
// -- currently just "us" -- are served on the dedicated REP host
// aiplatform.<jurisdiction>.rep.googleapis.com. An empty region returns "".
//
// See gcloud-python 579.0.0 release note "Route gcloud ai requests for the us
// multi-region to the Vertex AI multi-regional (REP) endpoint."
func aiplatformEndpointForRegion(region string) string {
	if region == "" {
		return ""
	}
	if isAIPlatformMREPJurisdiction(region) {
		return fmt.Sprintf("https://aiplatform.%s.rep.googleapis.com/", region)
	}
	return fmt.Sprintf("https://%s-aiplatform.googleapis.com/", region)
}

// isAIPlatformMREPJurisdiction reports whether the given region is a
// multi-regional (mREP) jurisdiction served on a dedicated REP host. This
// mirrors constants.MREP_JURISDICTIONS in gcloud-python.
func isAIPlatformMREPJurisdiction(region string) bool {
	return region == "us"
}

// ML returns a client for the legacy Cloud ML Engine API
// (google.golang.org/api/ml/v1), which backs the `gcloud ai-platform`
// surface. This is distinct from Vertex AI (`gcloud ai ...`), which uses
// the newer aiplatform.googleapis.com API and is region-scoped.
func (c *Clients) ML(ctx context.Context) (*ml.Service, error) {
	return newService(ctx, c, ml.NewService)
}

// PublicCA returns a client for the PublicCA API.
func (c *Clients) PublicCA(ctx context.Context) (*publicca.Service, error) {
	return newService(ctx, c, publicca.NewService)
}

// InfraManager returns a client for the InfraManager API.
func (c *Clients) InfraManager(ctx context.Context) (*config1.Service, error) {
	return newService(ctx, c, config1.NewService)
}

// AgentRegistry returns a client for the AgentRegistry API.
func (c *Clients) AgentRegistry(ctx context.Context) (*agentregistry.APIService, error) {
	return newService(ctx, c, agentregistry.NewService)
}

// AgentIdentity returns an agentidentity v1 client.
func (c *Clients) AgentIdentity(ctx context.Context) (*agentidentity.Service, error) {
	return newService(ctx, c, agentidentity.NewService)
}

// AgentIdentityAlpha returns an agentidentity v1alpha client. The
// v1alpha surface carries the auth-provider authorization IAM methods that
// v1 does not expose. A non-empty endpoint overrides the global base URL
// (used for --endpoint-mode=regional).
func (c *Clients) AgentIdentityAlpha(ctx context.Context, endpoint string) (*agentidentityalpha.Service, error) {
	var opts []option.ClientOption
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	return newService(ctx, c, agentidentityalpha.NewService, opts...)
}

// CloudIdentity returns a client for the CloudIdentity API.
func (c *Clients) CloudIdentity(ctx context.Context) (*cloudidentity.Service, error) {
	return newService(ctx, c, cloudidentity.NewService)
}

// NetworkConnectivity returns a client for the NetworkConnectivity API.
func (c *Clients) NetworkConnectivity(ctx context.Context) (*networkconnectivity.Service, error) {
	return newService(ctx, c, networkconnectivity.NewService)
}

// NetworkSecurity returns a client for the NetworkSecurity API.
func (c *Clients) NetworkSecurity(ctx context.Context) (*networksecurity.Service, error) {
	return newService(ctx, c, networksecurity.NewService)
}

// NetworkSecurityBeta returns a client for the NetworkSecurityBeta API.
func (c *Clients) NetworkSecurityBeta(ctx context.Context) (*networksecuritybeta.Service, error) {
	return newService(ctx, c, networksecuritybeta.NewService)
}

// SecurityCenter returns a Security Command Center V1 client.
func (c *Clients) SecurityCenter(ctx context.Context) (*securitycenter.Service, error) {
	return newService(ctx, c, securitycenter.NewService)
}

// withRequestLogging returns client options equivalent to opts, but with
// every outgoing request written to out as a curl command. It builds the
// transport the generated client would have built anyway, on top of a logging
// base, so what gets printed is the request as sent -- including the
// Authorization and X-Goog-User-Project headers added by the auth transport.
//
// A nil out returns opts unchanged, so callers can wrap unconditionally.
func withRequestLogging(ctx context.Context, out io.Writer, opts ...option.ClientOption) ([]option.ClientOption, error) {
	if out == nil {
		return opts, nil
	}
	client, err := loggingHTTPClient(ctx, out, opts...)
	if err != nil {
		return nil, err
	}
	return []option.ClientOption{option.WithHTTPClient(client)}, nil
}

// loggingHTTPClient builds the HTTP client that opts describe, with the
// logging transport spliced in underneath the auth transport.
func loggingHTTPClient(ctx context.Context, out io.Writer, opts ...option.ClientOption) (*http.Client, error) {
	trans, err := htransport.NewTransport(ctx, httplog.NewTransport(http.DefaultTransport, out), opts...)
	if err != nil {
		return nil, fmt.Errorf("building debug transport: %w", err)
	}
	return &http.Client{Transport: trans}, nil
}

// SecurityPosture returns a client for the SecurityPosture API.
func (c *Clients) SecurityPosture(ctx context.Context) (*securityposture.Service, error) {
	return newService(ctx, c, securityposture.NewService)
}

// DeploymentManager returns a client for the DeploymentManager API.
func (c *Clients) DeploymentManager(ctx context.Context) (*deploymentmanager.Service, error) {
	return newService(ctx, c, deploymentmanager.NewService)
}

// DeploymentManagerBeta returns a client for the DeploymentManagerBeta API.
func (c *Clients) DeploymentManagerBeta(ctx context.Context) (*deploymentmanagerbeta.Service, error) {
	return newService(ctx, c, deploymentmanagerbeta.NewService)
}

// CloudFunctionsV1 returns a client for the CloudFunctionsV1 API.
func (c *Clients) CloudFunctionsV1(ctx context.Context) (*cloudfunctionsv1.Service, error) {
	return newService(ctx, c, cloudfunctionsv1.NewService)
}

// CloudFunctionsV2 returns a client for the CloudFunctionsV2 API.
func (c *Clients) CloudFunctionsV2(ctx context.Context) (*cloudfunctionsv2.Service, error) {
	return newService(ctx, c, cloudfunctionsv2.NewService)
}

// AppEngine returns a client for the AppEngine API.
func (c *Clients) AppEngine(ctx context.Context) (*appengine.APIService, error) {
	return newService(ctx, c, appengine.NewService)
}

// Logging returns a client for the Logging API.
func (c *Clients) Logging(ctx context.Context) (*logging.Service, error) {
	return newService(ctx, c, logging.NewService)
}

// SQLAdmin returns a client for the SQLAdmin API.
func (c *Clients) SQLAdmin(ctx context.Context) (*sqladmin.Service, error) {
	return newService(ctx, c, sqladmin.NewService)
}

// StorageTransfer returns a client for the StorageTransfer API.
func (c *Clients) StorageTransfer(ctx context.Context) (*storagetransfer.Service, error) {
	return newService(ctx, c, storagetransfer.NewService)
}

// ServiceManagement returns a client for the ServiceManagement API.
func (c *Clients) ServiceManagement(ctx context.Context) (*servicemanagement.APIService, error) {
	return newService(ctx, c, servicemanagement.NewService)
}

// VMMigration returns a client for the VMMigration API.
func (c *Clients) VMMigration(ctx context.Context) (*vmmigration.Service, error) {
	return newService(ctx, c, vmmigration.NewService)
}

// SecureSourceManager returns a client for the SecureSourceManager API.
func (c *Clients) SecureSourceManager(ctx context.Context) (*securesourcemanager.Service, error) {
	return newService(ctx, c, securesourcemanager.NewService)
}

// Domains returns a client for the Domains API.
func (c *Clients) Domains(ctx context.Context) (*domains.Service, error) {
	return newService(ctx, c, domains.NewService)
}

// Filestore returns a client for the Filestore API.
func (c *Clients) Filestore(ctx context.Context) (*filestore.Service, error) {
	return newService(ctx, c, filestore.NewService)
}

// WorkflowExecutions returns a client for the WorkflowExecutions API.
func (c *Clients) WorkflowExecutions(ctx context.Context) (*workflowexecutions.Service, error) {
	return newService(ctx, c, workflowexecutions.NewService)
}

// Workflows returns a client for the Workflows API.
func (c *Clients) Workflows(ctx context.Context) (*workflows.Service, error) {
	return newService(ctx, c, workflows.NewService)
}

// MonitoringV1 returns a client for the MonitoringV1 API.
func (c *Clients) MonitoringV1(ctx context.Context) (*monitoringv1.Service, error) {
	return newService(ctx, c, monitoringv1.NewService)
}

// NetworkManagement returns a client for the NetworkManagement API.
func (c *Clients) NetworkManagement(ctx context.Context) (*networkmanagement.Service, error) {
	return newService(ctx, c, networkmanagement.NewService)
}

// CloudLocationFinder returns a client for the CloudLocationFinder API.
func (c *Clients) CloudLocationFinder(ctx context.Context) (*cloudlocationfinder.Service, error) {
	return newService(ctx, c, cloudlocationfinder.NewService)
}

// Datastore returns a client for the Datastore API.
func (c *Clients) Datastore(ctx context.Context) (*datastore.Service, error) {
	return newService(ctx, c, datastore.NewService)
}

// NetApp returns a client for the NetApp API.
func (c *Clients) NetApp(ctx context.Context) (*netapp.Service, error) {
	return newService(ctx, c, netapp.NewService)
}

// BigtableAdmin returns a client for the BigtableAdmin API.
func (c *Clients) BigtableAdmin(ctx context.Context) (*bigtableadmin.Service, error) {
	return newService(ctx, c, bigtableadmin.NewService)
}

// Composer returns a client for the Composer API.
func (c *Clients) Composer(ctx context.Context) (*composer.Service, error) {
	return newService(ctx, c, composer.NewService)
}

// ComposerBeta returns a Composer v1beta1 client. Environment
// hibernate/resume are only exposed on v1beta1 as of this build.
func (c *Clients) ComposerBeta(ctx context.Context) (*composerbeta.Service, error) {
	return newService(ctx, c, composerbeta.NewService)
}

// DataCatalog returns a client for the DataCatalog API.
func (c *Clients) DataCatalog(ctx context.Context) (*datacatalog.Service, error) {
	return newService(ctx, c, datacatalog.NewService)
}

// CloudDeploy returns a client for the CloudDeploy API.
func (c *Clients) CloudDeploy(ctx context.Context) (*clouddeploy.Service, error) {
	return newService(ctx, c, clouddeploy.NewService)
}

// DNS returns a client for the DNS API.
func (c *Clients) DNS(ctx context.Context) (*dns.Service, error) {
	return newService(ctx, c, dns.NewService)
}

// DNSAt returns a Cloud DNS v1 client pointed at endpoint (a
// regional endpoint base URL), or the global endpoint when endpoint is "".
func (c *Clients) DNSAt(ctx context.Context, endpoint string) (*dns.Service, error) {
	var opts []option.ClientOption
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint))
	}
	return newService(ctx, c, dns.NewService, opts...)
}

// ManagedIdentities returns a client for the ManagedIdentities API.
func (c *Clients) ManagedIdentities(ctx context.Context) (*managedidentities.Service, error) {
	return newService(ctx, c, managedidentities.NewService)
}

func (c *Clients) Dataproc(ctx context.Context, region string) (*dataproc.Service, error) {
	var opts []option.ClientOption
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-dataproc.googleapis.com/", region)))
	}
	return newService(ctx, c, dataproc.NewService, opts...)
}

// IDS returns a client for the IDS API.
func (c *Clients) IDS(ctx context.Context) (*ids.Service, error) {
	return newService(ctx, c, ids.NewService)
}

// ReCaptchaEnterprise returns a client for the ReCaptchaEnterprise API.
func (c *Clients) ReCaptchaEnterprise(ctx context.Context) (*recaptchaenterprise.Service, error) {
	return newService(ctx, c, recaptchaenterprise.NewService)
}

// ApiHub returns a client for the ApiHub API.
func (c *Clients) ApiHub(ctx context.Context) (*apihub.Service, error) {
	return newService(ctx, c, apihub.NewService)
}

// PubSubLite returns a pubsublite client wired to the region-specific
// endpoint. Pub/Sub Lite is a regional service: admin operations for a given
// location must be routed through <region>-pubsublite.googleapis.com, where
// the region is derived from the (regional or zonal) location the resource
// lives in.
func (c *Clients) PubSubLite(ctx context.Context, region string) (*pubsublite.Service, error) {
	var opts []option.ClientOption
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-pubsublite.googleapis.com/", region)))
	}
	return newService(ctx, c, pubsublite.NewService, opts...)
}

// ParameterManager returns a client for the ParameterManager API.
func (c *Clients) ParameterManager(ctx context.Context) (*parametermanager.Service, error) {
	return newService(ctx, c, parametermanager.NewService)
}

// RunV1 returns a Cloud Run v1 (Knative-style) client. Cloud Run's v1
// surface is regional; pass an empty region for the multi-region default
// endpoint or a region name (e.g. "us-central1") to route calls through
// https://REGION-run.googleapis.com/.
func (c *Clients) RunV1(ctx context.Context, region string) (*runv1.APIService, error) {
	var opts []option.ClientOption
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-run.googleapis.com/", region)))
	}
	return newService(ctx, c, runv1.NewService, opts...)
}

// RunV2 returns a Cloud Run v2 client. Cloud Run v2 supports both a
// global and per-region endpoint; pass a region to pin the call to
// https://REGION-run.googleapis.com/ (recommended for regional resources).
func (c *Clients) RunV2(ctx context.Context, region string) (*runv2.Service, error) {
	var opts []option.ClientOption
	if region != "" {
		opts = append(opts, option.WithEndpoint(fmt.Sprintf("https://%s-run.googleapis.com/", region)))
	}
	return newService(ctx, c, runv2.NewService, opts...)
}

// KMS returns a Cloud KMS v1 client for key rings, keys, versions,
// EKM connections, import jobs, key handles, retired resources, and
// single-tenant HSM management.
func (c *Clients) KMS(ctx context.Context) (*cloudkms.Service, error) {
	return newService(ctx, c, cloudkms.NewService)
}

// KMSInventory returns a KMS Inventory v1 client for listing keys and
// searching resources protected by a Cloud KMS key.
func (c *Clients) KMSInventory(ctx context.Context) (*kmsinventory.Service, error) {
	return newService(ctx, c, kmsinventory.NewService)
}

// Spanner returns a Cloud Spanner v1 client.
func (c *Clients) Spanner(ctx context.Context) (*spanner.Service, error) {
	return newService(ctx, c, spanner.NewService)
}

// PlatformTokenSource returns an OAuth token source with the cloud-platform
// scope for callers that need to make raw HTTP requests against Google Cloud
// endpoints not covered by a generated Go client (for example
// eventarcpublishing.googleapis.com).
func (c *Clients) PlatformTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	ts, err := c.Tokens.TokenSource(ctx, c.Account, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return ts, nil
}
