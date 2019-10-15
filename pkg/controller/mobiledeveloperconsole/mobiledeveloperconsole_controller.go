package mobiledeveloperconsole

import (
	"context"
	"fmt"

	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	"github.com/aerogear/mobile-developer-console-operator/pkg/config"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"

	"github.com/aerogear/mobile-developer-console-operator/pkg/util"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	openshiftappsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	cfg = config.New()
	log = logf.Log.WithName("controller_mobiledeveloperconsole")
)

const (
	ControllerName = "mobiledeveloperconsole-controller"
)

// Add creates a new MobileDeveloperConsole Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, autodetectChannel chan schema.GroupVersionKind) error {
	return add(mgr, newReconciler(mgr), autodetectChannel)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	client := mgr.GetClient()

	return &ReconcileMobileDeveloperConsole{
		client:  client,
		scheme:  mgr.GetScheme(),
		context: ctx,
		cancel:  cancel,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, autodetectChannel chan schema.GroupVersionKind) error {

	// Create a new controller
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &mdcv1alpha1.MobileDeveloperConsole{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccount and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Role and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RoleBinding and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Route and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource DeploymentConfig and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &openshiftappsv1.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ImageStream and requeue the owner MobileDeveloperConsole
	err = c.Watch(&source.Kind{Type: &imagev1.ImageStream{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})
	if err != nil {
		return err
	}

	go func() {
		for gvk := range autodetectChannel {
			// Check if this channel event was for the PrometheusRule resource type
			if gvk.String() == monitoringv1.SchemeGroupVersion.WithKind(monitoringv1.PrometheusRuleKind).String() {
				watchSecondaryResource(c, gvk, &monitoringv1.PrometheusRule{})
			}

			// Check if this channel event was for the ServiceMonitor resource type
			if gvk.String() == monitoringv1.SchemeGroupVersion.WithKind(monitoringv1.ServiceMonitorsKind).String() {
				watchSecondaryResource(c, gvk, &monitoringv1.ServiceMonitor{})
			}

			// Check if this channel event was for the GrafanaDashboard resource type
			if gvk.String() == integreatlyv1alpha1.SchemeGroupVersion.WithKind(integreatlyv1alpha1.GrafanaDashboardKind).String() {
				watchSecondaryResource(c, gvk, &integreatlyv1alpha1.GrafanaDashboard{})
			}
		}
	}()

	return nil
}

var _ reconcile.Reconciler = &ReconcileMobileDeveloperConsole{}

// ReconcileMobileDeveloperConsole reconciles a MobileDeveloperConsole object
type ReconcileMobileDeveloperConsole struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	context context.Context
	cancel  context.CancelFunc
}

// Reconcile reads the state of the cluster for a MobileDeveloperConsole object and makes changes based on the state read
// and what is in the MobileDeveloperConsole.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMobileDeveloperConsole) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Start Reconcile - MobileDeveloperConsole")

	// Fetch the MobileDeveloperConsole instance
	instance := &mdcv1alpha1.MobileDeveloperConsole{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.Phase == mdcv1alpha1.PhaseEmpty {
		instance.Status.Phase = mdcv1alpha1.PhaseProvision
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MDC resource status phase", "MDC.Namespace", instance.Namespace, "MDC.Name", instance.Name)
			return reconcile.Result{}, err
		}
	}

	//#region ServiceAccount
	serviceAccount, err := newMDCServiceAccount(instance)

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ServiceAccount already exists
	foundServiceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region MobileClientAdminRole
	role, err := newMobileClientAdminRole(instance)

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, role, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Role already exists
	foundRole := &rbacv1.Role{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Role", "Role.Namespace", role.Namespace, "Role.Name", role.Name)
		err = r.client.Create(context.TODO(), role)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region MobileClientAdminRoleBinding
	roleBinding, err := newMobileClientAdminRoleBinding(instance)

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, roleBinding, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this RoleBinding already exists
	foundRoleBinding := &rbacv1.RoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, foundRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new RoleBinding", "RoleBinding.Namespace", roleBinding.Namespace, "RoleBinding.Name", roleBinding.Name)
		err = r.client.Create(context.TODO(), roleBinding)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region DeleteOldRoleBinding
	// TODO: This can be removed after a release or two
	oldRoleBinding := &rbacv1.RoleBinding{}
	oldRoleBindingName := instance.Namespace + "-" + instance.Name + "-mobileclient-admin"
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oldRoleBindingName, Namespace: instance.Namespace}, oldRoleBinding)
	if err == nil && oldRoleBinding.Name != "" {
		err = r.client.Delete(context.TODO(), oldRoleBinding)
		if err != nil {
			reqLogger.Error(err, "Failed to delete old Rolebinding")
		}
	}
	//#endregion

	//#region OauthProxy Service
	oauthProxyService, err := newOauthProxyService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundOauthProxyService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyService.Name, Namespace: oauthProxyService.Namespace}, foundOauthProxyService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", oauthProxyService.Namespace, "Service.Name", oauthProxyService.Name)
		err = r.client.Create(context.TODO(), oauthProxyService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region MDC Service
	mdcService, err := newMDCService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, mdcService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundMDCService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcService.Name, Namespace: mdcService.Namespace}, foundMDCService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", mdcService.Namespace, "Service.Name", mdcService.Name)
		err = r.client.Create(context.TODO(), mdcService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region OauthProxy Route
	oauthProxyRoute, err := newOauthProxyRoute(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyRoute, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Route already exists
	foundOauthProxyRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyRoute.Name, Namespace: oauthProxyRoute.Namespace}, foundOauthProxyRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", oauthProxyRoute.Namespace, "Route.Name", oauthProxyRoute.Name)
		err = r.client.Create(context.TODO(), oauthProxyRoute)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region OauthProxy ImageStream
	oauthProxyImageStream, err := newOauthProxyImageStream(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyImageStream, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ImageStream already exists
	foundOauthProxyImageStream := &imagev1.ImageStream{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyImageStream.Name, Namespace: oauthProxyImageStream.Namespace}, foundOauthProxyImageStream)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ImageStream", "ImageStream.Namespace", foundOauthProxyImageStream.Namespace, "ImageStream.Name", oauthProxyImageStream.Name)
		err = r.client.Create(context.TODO(), oauthProxyImageStream)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region MDC ImageStream
	mdcImageStream, err := newMDCImageStream(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set MobileDeveloperConsole instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, mdcImageStream, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ImageStream already exists
	foundMDCImageStream := &imagev1.ImageStream{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcImageStream.Name, Namespace: mdcImageStream.Namespace}, foundMDCImageStream)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ImageStream", "ImageStream.Namespace", mdcImageStream.Namespace, "ImageStream.Name", mdcImageStream.Name)
		err = r.client.Create(context.TODO(), mdcImageStream)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region MDC DeploymentConfig
	mdcDeploymentConfig, err := newMDCDeploymentConfig(instance)

	if err := controllerutil.SetControllerReference(instance, mdcDeploymentConfig, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DeploymentConfig already exists
	foundMDCDeploymentConfig := &openshiftappsv1.DeploymentConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcDeploymentConfig.Name, Namespace: mdcDeploymentConfig.Namespace}, foundMDCDeploymentConfig)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DeploymentConfig", "DeploymentConfig.Namespace", mdcDeploymentConfig.Namespace, "DeploymentConfig.Name", mdcDeploymentConfig.Name)
		err = r.client.Create(context.TODO(), mdcDeploymentConfig)
		if err != nil {
			return reconcile.Result{}, err
		}

		// DeploymentConfig created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region Monitoring
	//## region ServiceMonitor
	mdcServiceMonitor, err := newMDCServiceMonitor(instance)
	if err := controllerutil.SetControllerReference(instance, mdcServiceMonitor, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this Service Monitor already exists
	foundMDCServiceMonitor := &monitoringv1.ServiceMonitor{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcServiceMonitor.Name, Namespace: mdcServiceMonitor.Namespace}, foundMDCServiceMonitor)
	if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
		reqLogger.Info(fmt.Sprintf("%s is not available on the cluster, the resource wont be created moving on. Install prometheus-operator in you cluster to create %s objects", kindMatchErr.GroupKind, kindMatchErr.GroupKind))
	} else {
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new ServiceMonitor", "ServiceMonitor.Namespace", mdcServiceMonitor.Namespace, "ServiceMonitor.Name", mdcServiceMonitor.Name)
			err = r.client.Create(context.TODO(), mdcServiceMonitor)
			if err != nil {
				return reconcile.Result{}, err
			}
			// ServiceMonitor created successfully
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	//## endregion ServiceMonitor

	//## region PrometheusRule

	mdcPrometheusRule, err := newMDCPrometheusRule(instance)

	if err := controllerutil.SetControllerReference(instance, mdcPrometheusRule, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Prometheus Rule already exists
	foundMDCPrometheusRule := &monitoringv1.PrometheusRule{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcPrometheusRule.Name, Namespace: mdcPrometheusRule.Namespace}, foundMDCPrometheusRule)
	if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
		reqLogger.Info(fmt.Sprintf("%s is not available on the cluster, the resource wont be created moving on. Install prometheus-operator in you cluster to create %s objects", kindMatchErr.GroupKind, kindMatchErr.GroupKind))
	} else {
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new PrometheusRule", "PrometheusRule.Namespace", mdcPrometheusRule.Namespace, "PrometheusRule.Name", mdcPrometheusRule.Name)
			err = r.client.Create(context.TODO(), mdcPrometheusRule)
			if err != nil {
				return reconcile.Result{}, err
			}
			// PrometheusRule created successfully
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}
	//## endregion PrometheusRule

	//## region GrafanaDasboard
	mdcGrafanaDashboard, err := newMDCGrafanaDashboard(instance)

	if err := controllerutil.SetControllerReference(instance, mdcGrafanaDashboard, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Grafana Dasboard already exists
	foundMDCGrafanaDashboard := &integreatlyv1alpha1.GrafanaDashboard{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: mdcGrafanaDashboard.Name, Namespace: mdcGrafanaDashboard.Namespace}, foundMDCGrafanaDashboard)
	if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
		reqLogger.Info(fmt.Sprintf("%s is not available on the cluster, the resource wont be created moving on. Install grafana-operator in you cluster to create %s objects", kindMatchErr.GroupKind, kindMatchErr.GroupKind))
	} else {
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new GrafanaDashboard", "GrafanaDashboard.Namespace", mdcGrafanaDashboard.Namespace, "GrafanaDashboard.Name", mdcGrafanaDashboard.Name)
			err = r.client.Create(context.TODO(), mdcGrafanaDashboard)
			if err != nil {
				return reconcile.Result{}, err
			}
			// GrafanaDasboard created successfully
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	//## endregion GrafanaDasboard
	//#endregion

	if foundMDCDeploymentConfig.Status.ReadyReplicas > 0 && instance.Status.Phase != mdcv1alpha1.PhaseComplete {
		instance.Status.Phase = mdcv1alpha1.PhaseComplete
		r.client.Status().Update(context.TODO(), instance)
	}
	// Resources already exist - don't requeue
	reqLogger.Info("Finish Reconcile - MobileDeveloperConsole")
	return reconcile.Result{}, nil
}

func watchSecondaryResource(c controller.Controller, gvk schema.GroupVersionKind, o runtime.Object) error {
	stateManager := util.GetStateManager()
	stateFieldName := getStateFieldName(gvk.Kind)

	// Check to see if the watch exists for this resource type already for this controller, if it does, we return so we don't set up another watch
	watchExists, keyExists := stateManager.GetState(stateFieldName).(bool)
	if keyExists || watchExists {
		return nil
	}

	// Set up the actual watch
	err := c.Watch(&source.Kind{Type: o}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mdcv1alpha1.MobileDeveloperConsole{},
	})

	// Retry on error
	if err != nil {
		log.Error(err, "error creating watch")
		stateManager.SetState(stateFieldName, false)
		return err
	}

	stateManager.SetState(stateFieldName, true)
	log.Info(fmt.Sprintf("Watch created for '%s' resource in '%s'", gvk.Kind, ControllerName))
	return nil
}

func getStateFieldName(kind string) string {
	return ControllerName + "-watch-" + kind
}
