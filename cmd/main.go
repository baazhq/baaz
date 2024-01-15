package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"github.com/gorilla/handlers"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	khota "datainfra.io/baaz/internal/khota_handler"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	datainfraiov1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/internal/app_controller"
	dataplane_controller "datainfra.io/baaz/internal/dataplane_controller"
	tenant_controller "datainfra.io/baaz/internal/tenant_controller"
	tenantinfra_controller "datainfra.io/baaz/internal/tenantinfra_controller"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(datainfraiov1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// saas initalizer configures ports so that local development
// in a single host machine does not conflict for shared, dedciated
// and private saas.
type saasInitalizer struct {
	CustomerName     string
	HttpServerPort   string
	HealthProbePort  string
	MetricServerPort string
}

func newSaaSinitalizer(enablePrivateSaaS, runLocal bool) *saasInitalizer {

	if enablePrivateSaaS {
		return &saasInitalizer{
			HealthProbePort:  ":7001",
			MetricServerPort: ":7002",
		}
	}

	return &saasInitalizer{
		HttpServerPort:   ":8000",
		HealthProbePort:  ":8001",
		MetricServerPort: ":8002",
	}
}

func main() {
	var enableLeaderElection bool
	var enablePrivateSaaS bool
	var customerName string
	var runLocal bool

	flag.BoolVar(&enablePrivateSaaS, "private_mode", false, "Enable private mode runs BaaZ controllers in a private saas mode.")
	flag.StringVar(&customerName, "customer_name", "", "Customer name for private saas")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. "+"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	saasInit := newSaaSinitalizer(enablePrivateSaaS, runLocal)

	if !enablePrivateSaaS {
		go func() {
			router := khota.NewRouter()
			setupLog.Info(fmt.Sprintf("Started BaaZ HTTP server on :%s", saasInit.HttpServerPort))
			if err := http.ListenAndServe(saasInit.HttpServerPort, handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Access-Control-Allow-Origin"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "DELETE", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)); err != nil {
				setupLog.Error(err, "unable to start HTTP server")
				os.Exit(1)
			}
		}()
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: saasInit.HealthProbePort,
		LeaderElection:         enableLeaderElection,
		Metrics: server.Options{
			BindAddress: saasInit.MetricServerPort,
		},
		LeaderElectionID: "72b9bc85.datainfra.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (dataplane_controller.NewDataplaneReconciler(mgr, enablePrivateSaaS, customerName)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Dataplane")
		os.Exit(1)
	}

	if err = (app_controller.NewApplicationReconciler(mgr, enablePrivateSaaS, customerName)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	if err = (tenantinfra_controller.NewTenantsInfraReconciler(mgr, enablePrivateSaaS, customerName)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TenantInfra")
		os.Exit(1)
	}

	if err = (tenant_controller.NewTenantsReconciler(mgr, enablePrivateSaaS, customerName)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
