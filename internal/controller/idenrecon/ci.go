/*
 Copyright 2022 The ANKA SOFTWARE Authors.
*/

package idenrecon

import (
	"context"
	"fmt"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sdkIdenRecon "github.com/anka-software/cmdb-sdk/pkg/client/cmdb"

	sdkMeta "github.com/anka-software/cmdb-sdk/pkg/client/cmdb_meta"
	sdkTable "github.com/anka-software/cmdb-sdk/pkg/client/table"
	"github.com/crossplane/provider-cmdb/apis/idenrecon/v1alpha1"
	apisv1alpha1 "github.com/crossplane/provider-cmdb/apis/v1alpha1"
	"github.com/crossplane/provider-cmdb/internal/clients"
	"github.com/crossplane/provider-cmdb/internal/clients/idenrecon"
	cmdbmeta "github.com/crossplane/provider-cmdb/internal/clients/meta"
	"github.com/crossplane/provider-cmdb/internal/clients/table"
	"github.com/crossplane/provider-cmdb/internal/controller/features"
)

const (
	errNotCI        = "managed resource is not a CI custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"

	errCreateFailed = "cannot create CI with Identification and Reconciliation API"
	// errGetFailed    = "cannot get CI with Table API"
	// errDeleteFailed = "cannot delete CI with Table API"
)

// Setup adds a controller that reconciles Identification and Reconciliation managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.CIGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CIGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:                  mgr.GetClient(),
			usage:                 resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFnIdenRecon: idenrecon.NewIdenReconClient,
			newServiceFnTable:     table.NewTableClient,
			newServiceFnMeta:      cmdbmeta.NewMetaClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.CI{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube                  client.Client
	usage                 resource.Tracker
	newServiceFnIdenRecon func(cfg clients.Config) sdkIdenRecon.ClientService
	newServiceFnTable     func(cfg clients.Config) sdkTable.ClientService
	newServiceFnMeta      func(cfg clients.Config) sdkMeta.ClientService
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CI)
	if !ok {
		return nil, errors.New(errNotCI)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{kube: c.kube, serviceIdenRecon: c.newServiceFnIdenRecon(*cfg), serviceTable: c.newServiceFnTable(*cfg), serviceMeta: c.newServiceFnMeta(*cfg)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube client.Client
	// A 'client' used to connect to the external resource API.
	serviceIdenRecon sdkIdenRecon.ClientService
	serviceTable     sdkTable.ClientService
	serviceMeta      sdkMeta.ClientService
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CI)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCI)
	}

	// These fmt statements should be removed in the real implementation.
	fmt.Printf("Observing: \n%+v", cr)

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	forProvider := &cr.Spec.ForProvider
	params := table.GenerateGetTableItemsOptions(forProvider.ClassName, forProvider.Name)

	/*if forProvider.Name == ""{
		return managed.ExternalObservation{ResourceExists: false}, nil
	}*/

	desired := cr.Spec.ForProvider.DeepCopy()

	response, err := c.serviceTable.GetTableItems(params)

	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if len(response.Payload.Result) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	metaParams := cmdbmeta.GenerateGetMetaOptions(desired.ClassName)
	responseMeta, err := c.serviceMeta.GetCmdbMetaByClassName(metaParams)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(err.Error())

	}

	var elementNames []string
	for _, v := range responseMeta.Payload.Result.Attributes {
		elementNames = append(elementNames, v.Element) // # CHANGED
		//elementNames[i] = v.Element
	}

	err = idenrecon.ContainsField(elementNames, desired.Values)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(err.Error())
	}

	currentResource := response.Payload.Result[0]

	resourceUpToDate := idenrecon.IsResourceUpToDate(desired.Values, currentResource)
	// currentResource

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: resourceUpToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CI)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCI)
	}

	fmt.Printf("Creating: \n%+v", cr)

	cr.Status.SetConditions(xpv1.Creating())

	ciParams := idenrecon.GenerateCIOptions(&cr.Spec.ForProvider)

	response, err := c.serviceIdenRecon.CreateIdentifyReconcile(ciParams)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	var item = (*response.Payload.Result.Items)[0]
	fmt.Println("Sys ID : " + item.SysId)

	meta.SetExternalName(cr, item.SysId)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CI)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCI)
	}

	fmt.Printf("Updating: \n%+v", cr)

	cr.Status.SetConditions(xpv1.Creating())

	ciParams := idenrecon.GenerateCIOptions(&cr.Spec.ForProvider)

	response, err := c.serviceIdenRecon.CreateIdentifyReconcile(ciParams)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateFailed)
	}

	var item = (*response.Payload.Result.Items)[0]
	fmt.Println("Sys ID : " + item.SysId)

	// meta.SetExternalName(cr, item.SysId)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CI)
	if !ok {
		return errors.New(errNotCI)
	}

	fmt.Printf("Deleting: \n%+v", cr)

	cr.Status.SetConditions(xpv1.Deleting())

	return nil
}
