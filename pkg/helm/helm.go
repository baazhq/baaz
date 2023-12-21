package helm

import (
	"context"
	"fmt"
	"io/ioutil"
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
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var settings *cli.EnvSettings

type HelmAct interface {
	Apply(rest *rest.Config) error
	Uninstall(rest *rest.Config) error
	List(rest *rest.Config) (status string, exists bool)
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
func (h *Helm) List(rest *rest.Config) (status string, exists bool) {

	settings := cli.New()
	restGetter := NewRESTClientGetter(rest, h.Namespace)

	if err := h.Action.Init(&restGetter, h.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		return "", false
	}

	clientList := action.NewList(h.Action)

	settings.EnvVars()

	// Only list deployed
	clientList.Deployed = true
	results, err := clientList.Run()
	if err != nil {
		return "", false
	}

	for _, result := range results {
		if result.Name == h.ReleaseName {
			return result.Info.Status.String(), true
		}
	}
	return "", false
}

func (h *Helm) Uninstall(rest *rest.Config) error {

	settings := cli.New()

	restGetter := NewRESTClientGetter(rest, h.Namespace)

	if err := h.Action.Init(&restGetter, h.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		return err
	}

	client := action.NewUninstall(h.Action)

	settings.EnvVars()

	client.Wait = true
	client.Timeout = 120 * time.Second

	release, err := client.Run(h.ChartName)
	if err != nil {
		return err
	}

	klog.Infof("Uninstalling Release Name: [%s] Namespace [%s] Status [%s]", release.Release.Name, release.Release.Namespace, release.Release.Info.Status)

	return nil
}

// HelmInstall Method installs the chart.
// ref: https://github.com/PrasadG193/helm-clientgo-example/tree/master
func (h *Helm) Apply(rest *rest.Config) error {

	settings := cli.New()

	restGetter := NewRESTClientGetter(rest, h.Namespace)

	if err := h.Action.Init(&restGetter, h.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		return err
	}

	client := action.NewInstall(h.Action)

	settings.EnvVars()

	repoAdd(h.RepoName, h.RepoUrl)

	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", h.RepoName, h.ChartName), settings)
	if err != nil {
		return err
	}

	err = h.RepoUpdate()
	if err != nil {
		return err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return err
	}

	client.ReleaseName = h.ReleaseName
	client.Namespace = h.Namespace
	client.CreateNamespace = true
	client.Wait = true
	client.Timeout = 120 * time.Second

	client.WaitForJobs = true
	//client.IncludeCRDs = true

	values := values.Options{
		Values: h.Values,
	}

	vals, err := values.MergeValues(getter.All(settings))
	if err != nil {
		return err
	}

	release, err := client.Run(chartRequested, vals)
	if err != nil {
		return err
	}

	klog.Infof("Release Name: [%s] Namespace [%s] Status [%s]", release.Name, release.Namespace, release.Info.Status)

	return nil
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

	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				klog.Error("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
				return
			}
		}(re)
	}
	wg.Wait()

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
		klog.Error(err)
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
		klog.Error(err)
	}

	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		klog.Error(err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		klog.Error(err)
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
		klog.Error(err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		klog.Error(err)
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		klog.Error(err)
	}
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
