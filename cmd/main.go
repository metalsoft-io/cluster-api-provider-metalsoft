/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	cgrecord "k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/flags"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	infrav1alpha1 "github.com/metalsoft-io/cluster-api-provider-metalsoft/api/v1alpha1"
	controllers "github.com/metalsoft-io/cluster-api-provider-metalsoft/internal/controller"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/util/reconciler"
	"github.com/metalsoft-io/cluster-api-provider-metalsoft/version"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"

	//+kubebuilder:scaffold:imports

	metalsoft "github.com/metalsoft-io/cluster-api-provider-metalsoft/pkg/cloud/metalsoft"
)

var (
	scheme     = runtime.NewScheme()
	logOptions = logs.NewOptions()
	setupLog   = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(infrav1alpha1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

var (
	enableLeaderElection        bool
	leaderElectionLeaseDuration time.Duration
	leaderElectionRenewDeadline time.Duration
	leaderElectionRetryPeriod   time.Duration
	leaderElectionNamespace     string
	healthAddr                  string
	metricsAddr                 string
	profilerAddress             string
	metalsoftClusterConcurrency int
	metalsoftMachineConcurrency int
	webhookPort                 int
	webhookCertDir              string
	tlsOptions                  = flags.TLSOptions{}
	watchNamespace              string
	watchFilterValue            string
	syncPeriod                  time.Duration
	reconcileTimeout            time.Duration
	restConfigQPS               float32
	restConfigBurst             int
	showVersion                 bool
	// scopeCacheMaxSize           int
)

// InitFlags initializes the flags.
func InitFlags(fs *pflag.FlagSet) {
	logsv1.AddFlags(logOptions, fs)

	fs.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	fs.DurationVar(&leaderElectionLeaseDuration, "leader-elect-lease-duration",
		15*time.Second,
		"Interval at which non-leader candidates will wait to force acquire leadership (duration string)")

	fs.DurationVar(&leaderElectionRenewDeadline, "leader-elect-renew-deadline",
		10*time.Second,
		"Duration that the leading controller manager will retry refreshing leadership before giving up (duration string)")

	fs.DurationVar(&leaderElectionRetryPeriod, "leader-elect-retry-period",
		2*time.Second,
		"Duration the LeaderElector clients should wait between tries of actions (duration string)")

	fs.StringVar(&leaderElectionNamespace, "leader-election-namespace", "",
		"Namespace that the controller performs leader election in. If unspecified, the controller will discover which namespace it is running in.")

	fs.StringVar(&healthAddr, "health-addr", ":9440",
		"The address the health endpoint binds to.")

	fs.StringVar(&metricsAddr, "metrics-bind-address", "localhost:8080",
		"The address the metric endpoint binds to.")

	fs.StringVar(&profilerAddress, "profiler-address", "",
		"Bind address to expose the pprof profiler (e.g. localhost:6060)")

	fs.IntVar(&metalsoftClusterConcurrency, "metalsoftcluster-concurrency",
		10,
		"Number of MetalsoftClusters to process simultaneously")

	fs.IntVar(&metalsoftMachineConcurrency, "metalsoftmachine-concurrency",
		10,
		"Number of MetalsoftMachines to process simultaneously")

	fs.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook Server port")

	fs.StringVar(&webhookCertDir, "webhook-cert-dir",
		"/tmp/k8s-webhook-server/serving-certs/",
		"Webhook cert dir, only used when webhook-port is specified.")

	flags.AddTLSOptions(fs, &tlsOptions)

	fs.StringVar(&watchNamespace, "namespace", "",
		"Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.")

	fs.StringVar(&watchFilterValue, "watch-filter", "",
		fmt.Sprintf("Label value that the controller watches to reconcile cluster-api objects. Label key is always %s. If unspecified, the controller watches for all cluster-api objects.", clusterv1.WatchLabel))

	fs.DurationVar(&syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")

	fs.DurationVar(&reconcileTimeout, "reconcile-timeout",
		reconciler.DefaultLoopTimeout,
		"The maximum duration a reconcile loop can run (e.g. 90m)")

	fs.Float32Var(&restConfigQPS, "kube-api-qps", 20,
		"Maximum queries per second from the controller client to the Kubernetes API server. Defaults to 20")

	fs.IntVar(&restConfigBurst, "kube-api-burst", 30,
		"Maximum number of queries that should be allowed in one burst from the controller client to the Kubernetes API server. Default 30")

	fs.BoolVar(&showVersion, "version", false, "Show current version and exit.")
}

func main() {
	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if showVersion {
		fmt.Println(version.Get().String()) //nolint:forbidigo
		os.Exit(0)
	}

	if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	tlsOptionOverrides, err := flags.GetTLSOptionOverrideFuncs(tlsOptions)
	if err != nil {
		setupLog.Error(err, "unable to add TLS settings to the webhook server")
		os.Exit(1)
	}

	var watchNamespaces map[string]cache.Config
	if watchNamespace != "" {
		watchNamespaces = map[string]cache.Config{
			watchNamespace: {},
		}
	}

	if profilerAddress != "" {
		setupLog.Info(fmt.Sprintf("Profiler listening for requests at %s", profilerAddress))
		go func() {
			server := &http.Server{
				Addr: profilerAddress,

				// Timeouts
				ReadTimeout:       60 * time.Second,
				ReadHeaderTimeout: 60 * time.Second,
				WriteTimeout:      60 * time.Second,
				IdleTimeout:       60 * time.Second,
			}
			err := server.ListenAndServe()
			if err != nil {
				setupLog.Error(err, "listen and serve error")
			}
		}()
	}

	// klog.Background will automatically use the right logger.
	ctrl.SetLogger(klog.Background())

	// Machine and cluster operations can create enough events to trigger the event recorder spam filter
	// Setting the burst size higher ensures all events will be recorded and submitted to the API
	broadcaster := cgrecord.NewBroadcasterWithCorrelatorOptions(cgrecord.CorrelatorOptions{
		BurstSize: 100,
	})

	restConfig := ctrl.GetConfigOrDie()
	restConfig.QPS = restConfigQPS
	restConfig.Burst = restConfigBurst

	req, _ := labels.NewRequirement(clusterv1.ClusterNameLabel, selection.Exists, nil)
	clusterSecretCacheSelector := labels.NewSelector().Add(*req)

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                     scheme,
		Metrics:                    metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:     healthAddr,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "controller-leader-election-capms",
		LeaderElectionNamespace:    leaderElectionNamespace,
		LeaseDuration:              &leaderElectionLeaseDuration,
		RenewDeadline:              &leaderElectionRenewDeadline,
		RetryPeriod:                &leaderElectionRetryPeriod,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		EventBroadcaster:           broadcaster,
		Cache: cache.Options{
			DefaultNamespaces: watchNamespaces,
			SyncPeriod:        &syncPeriod,
			ByObject: map[client.Object]cache.ByObject{
				// Note: Only Secrets with the cluster name label are cached.
				// The default client of the manager won't use the cache for secrets at all (see Client.Cache.DisableFor).
				// The cached secrets will only be used by the secretCachingClient we create below.
				&corev1.Secret{}: {
					Label: clusterSecretCacheSelector,
				},
			},
		},
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&corev1.ConfigMap{},
					&corev1.Secret{},
				},
			},
		},
		WebhookServer: webhook.NewServer(
			webhook.Options{
				Port:    webhookPort,
				CertDir: webhookCertDir,
				TLSOpts: tlsOptionOverrides,
			},
		),

		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("metalsoft-controller"))

	// Setup the context that's going to be used in controllers and for the manager.
	ctx := ctrl.SetupSignalHandler()

	if setupErr := setupReconcilers(ctx, mgr); setupErr != nil {
		setupLog.Error(err, "unable to setup reconcilers")
		os.Exit(1)
	}

	if setupErr := setupWebhooks(mgr); setupErr != nil {
		setupLog.Error(err, "unable to setup webhooks")
		os.Exit(1)
	}

	if setupErr := setupProbes(mgr); setupErr != nil {
		setupLog.Error(err, "unable to setup probes")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	setupLog.Info("starting manager", "version", version.Get().String())
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupReconcilers(ctx context.Context, mgr ctrl.Manager) error {
	client, err := metalsoft.GetClient()
	if err != nil {
		setupLog.Error(err, "unable to create Metalsoft client")
		os.Exit(1)
	}

	if err = (&controllers.MetalsoftClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		// Recorder:         mgr.GetEventRecorderFor("metalsoft-controller"),
		MetalSoftClient:  client,
		ReconcileTimeout: reconcileTimeout,
		WatchFilterValue: watchFilterValue,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: metalsoftClusterConcurrency}); err != nil {
		return fmt.Errorf("setting up MetalsoftCluster controller: %w", err)
	}

	if err = (&controllers.MetalsoftMachineReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		// Recorder:         mgr.GetEventRecorderFor("metalsoft-controller"),
		MetalSoftClient:  client,
		ReconcileTimeout: reconcileTimeout,
		WatchFilterValue: watchFilterValue,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: metalsoftMachineConcurrency}); err != nil {
		return fmt.Errorf("setting up MetalsoftMachine controller: %w", err)
	}

	return nil
}

func setupWebhooks(mgr ctrl.Manager) error {
	if err := (&infrav1alpha1.MetalsoftCluster{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("setting up MetalsoftCluster webhook: %w", err)
	}
	if err := (&infrav1alpha1.MetalsoftClusterTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("setting up MetalsoftClusterTemplate webhook: %w", err)
	}
	if err := (&infrav1alpha1.MetalsoftMachine{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("setting up MetalsoftMachine webhook: %w", err)
	}
	if err := (&infrav1alpha1.MetalsoftMachineTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("setting up MetalsoftMachineTemplate webhook: %w", err)
	}

	return nil
}

func setupProbes(mgr ctrl.Manager) error {
	if err := mgr.AddReadyzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		return fmt.Errorf("creating ready check: %w", err)
	}

	if err := mgr.AddHealthzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		return fmt.Errorf("creating health check: %w", err)
	}

	return nil
}
