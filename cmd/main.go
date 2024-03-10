package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"github.com/gorilla/handlers"
	"github.com/parseablehq/parseable-sdk-go/parseable"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	khota "github.com/baazhq/baaz/internal/khota_handler"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	datainfraiov1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/internal/app_controller"
	dataplane_controller "github.com/baazhq/baaz/internal/dataplane_controller"
	tenant_controller "github.com/baazhq/baaz/internal/tenant_controller"
	tenantinfra_controller "github.com/baazhq/baaz/internal/tenantinfra_controller"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	streams  = []string{
		"customers",
		"dataplanes",
		"tenantsinfra",
		"tenants",
		"applications",
	}
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

func newSaaSinitalizer(enablePrivateSaaS bool) *saasInitalizer {

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

	flag.BoolVar(&enablePrivateSaaS, "private_mode", false, "Enable private mode runs BaaZ controllers in a private saas mode.")
	flag.StringVar(&customerName, "customer_name", "", "Customer name for private saas")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. "+"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	saasInit := newSaaSinitalizer(enablePrivateSaaS)

	if !enablePrivateSaaS {
		go func() {
			router := khota.NewRouter()
			setupLog.Info(fmt.Sprintf("started baaz http server on :%s", saasInit.HttpServerPort))
			if err := http.ListenAndServe(saasInit.HttpServerPort, handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Access-Control-Allow-Origin"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "DELETE", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)); err != nil {
				setupLog.Error(err, "unable to start http server")
				os.Exit(1)
			}
		}()
	}

	if os.Getenv("PARSEABLE_ENABLE") == "true" {
		for _, stream := range streams {
			createStream := parseable.NewStreamBuilder(
				stream,
				nil,
				nil,
				nil,
			)

			resp, err := createStream.CreateStream()
			if err != nil && resp == 400 {
				setupLog.Error(err, "create stream failed for parseable")
				os.Exit(1)
			}
		}
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
