package helm

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var settings *cli.EnvSettings

type HelmAct interface {
	HelmInstall(rest *rest.Config) error
	HelmList(rest *rest.Config) error
}

type Helm struct {
	Action      *action.Configuration
	ReleaseName string
	Namespace   string
	Values      []string
	RepoName    string
	ChartName   string
	RepoUrl     string
}

func NewHelm(
	releaseName, namespace, chartName, repoName, repoUrl string,
	values []string) HelmAct {
	return &Helm{
		Action:      new(action.Configuration),
		ReleaseName: releaseName,
		Namespace:   namespace,
		RepoName:    repoName,
		RepoUrl:     repoUrl,
		ChartName:   chartName,
		Values:      values,
	}
}

// HelmList Method installs the chart.
// https://helm.sh/docs/topics/advanced/#simple-example
func (h *Helm) HelmList(rest *rest.Config) error {

	settings := cli.New()
	restGetter := NewRESTClientGetter(rest, h.Namespace)

	if err := h.Action.Init(&restGetter, h.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		return err
	}

	clientList := action.NewList(h.Action)

	settings.EnvVars()

	// Only list deployed
	clientList.Deployed = true
	results, err := clientList.Run()
	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Name == h.ReleaseName {
			return errors.New(fmt.Sprintf("Err: Helm Chart exists [%s]", h.ReleaseName))
		}
	}
	return err
}

// HelmInstall Method installs the chart.
// ref: https://github.com/PrasadG193/helm-clientgo-example/tree/master
func (h *Helm) HelmInstall(rest *rest.Config) error {

	settings := cli.New()

	repoAdd(h.RepoName, h.RepoUrl)

	restGetter := NewRESTClientGetter(rest, h.Namespace)

	if err := h.Action.Init(&restGetter, h.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		return err
	}

	client := action.NewInstall(h.Action)

	settings.EnvVars()

	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", h.RepoName, h.ChartName), settings)
	if err != nil {
		return err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		log.Fatal(err)
	}

	client.ReleaseName = h.ReleaseName
	client.Namespace = h.Namespace
	client.CreateNamespace = true
	client.Wait = true
	client.Timeout = 120 * time.Second

	client.WaitForJobs = true
	client.IncludeCRDs = true

	values := values.Options{
		ValueFiles: h.Values,
	}

	vals, err := values.MergeValues(getter.All(settings))
	if err != nil {
		return err
	}

	release, err := client.Run(chartRequested, vals)
	if err != nil {
		log.Printf("%+v", err)
	}

	klog.Infof("Release Name: [%s] Namespace [%s] Status [%s]", release.Name, release.Namespace, release.Info.Status)

	return nil
}

// ref: https://github.com/PrasadG193/helm-clientgo-example/tree/master
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// ref: https://github.com/PrasadG193/helm-clientgo-example/tree/master
// RepoUpdate updates charts for all helm repos
func (helm *Helm) RepoUpdate() error {
	settings = cli.New()

	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		klog.Error(errors.New("no repositories found. You must add one before updating"))
		return err
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}

	klog.Info("Hang tight while we grab the latest from your chart repositories...\n")
	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				klog.Error("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
				return
			} else {
				klog.Info("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	klog.Info("Update Complete. ⎈ Happy Helming!⎈\n")

	return nil
}

// Ref: https://github.com/PrasadG193/helm-clientgo-example/tree/master
// RepoAdd adds repo with given name and url
func repoAdd(name, url string) {
	settings := cli.New()

	repoFile := settings.RepositoryConfig

	//Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		klog.Fatal(err)
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		klog.Fatal(err)
	}

	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		klog.Fatal(err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		klog.Fatal(err)
	}

	if f.Has(name) {
		return
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		klog.Fatal(err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		klog.Fatal(err)
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		klog.Fatal(err)
	}
	klog.Infof("%q has been added to your repositories\n", name)
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	klog.InfoDepth(2, fmt.Sprintf(format, v...))
}

type simpleRESTClientGetter struct {
	config    *rest.Config
	namespace string
}

func NewRESTClientGetter(config *rest.Config, namespace string) simpleRESTClientGetter {
	return simpleRESTClientGetter{
		namespace: namespace,
		config:    config,
	}
}

func (c *simpleRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return c.config, nil
}

func (c *simpleRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	config.Burst = 100

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)
	return memory.NewMemCacheClient(discoveryClient), nil
}

func (c *simpleRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (c *simpleRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.Context.Namespace = c.namespace

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
