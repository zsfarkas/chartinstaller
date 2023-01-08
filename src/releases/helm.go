package releases

import (
	"log"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func getReleaseNames(targetNamespace string) (*[]string, error) {
	actionConfig, _, err := createActionConfig(targetNamespace)
	if err != nil {
		return nil, err
	}

	client := action.NewList(actionConfig)
	results, err := client.Run()
	if err != nil {
		return nil, err
	}

	releaseList := make([]string, 0)

	for _, rel := range results {
		releaseList = append(releaseList, rel.Name)
	}

	return &releaseList, nil
}

func getReleaseStatus(releaseName string, targetNamespace string) (*ReleaseStatus, error) {
	actionConfig, _, err := createActionConfig(targetNamespace)
	if err != nil {
		return nil, err
	}

	client := action.NewStatus(actionConfig)
	rel, err := client.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return NewReleaseStatusFrom(rel), err
}

func uninstallRelease(releaseName string, targetNamespace string) (*ReleaseStatus, error) {
	actionConfig, _, err := createActionConfig(targetNamespace)
	if err != nil {
		return nil, err
	}

	client := action.NewUninstall(actionConfig)
	response, err := client.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return NewReleaseStatusFrom(response.Release), err
}

func installOrUpgradeRelease(releaseName string, repoUrl string, chartName string, values map[string]interface{}, targetNamespace string) (*ReleaseStatus, error) {
	actionConfig, settings, err := createActionConfig(targetNamespace)
	if err != nil {
		return nil, err
	}

	var chartPathOptions action.ChartPathOptions = action.ChartPathOptions{
		RepoURL: repoUrl,
	}

	chart, err := getChart(chartPathOptions, chartName, settings)
	if err != nil {
		return nil, err
	}

	var rel *release.Release
	var requestError error

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
		clientInstall := action.NewInstall(actionConfig)
		clientInstall.ReleaseName = releaseName
		clientInstall.Namespace = targetNamespace
		clientInstall.ChartPathOptions = chartPathOptions

		rel, requestError = clientInstall.Run(chart, values)
	} else {
		clientUpgrade := action.NewUpgrade(actionConfig)
		clientUpgrade.Namespace = targetNamespace
		clientUpgrade.ChartPathOptions = chartPathOptions

		rel, requestError = clientUpgrade.Run(releaseName, chart, values)
	}

	return NewReleaseStatusFrom(rel), requestError
}

func initRepo(repoUrl string, targetNamespace string) error {
	log.Printf("initializing repository '%s'...", repoUrl)

	entry := repo.Entry{
		URL:  repoUrl,
		Name: "repo",
	}

	_, settings, err := createActionConfig(targetNamespace)
	if err != nil {
		return err
	}

	log.Printf("creating dirs for repo file...")
	err = os.MkdirAll(filepath.Dir(settings.RepositoryConfig), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	log.Printf("loading repo file...")
	repoFile, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		log.Printf("could not load repo file, trying to create...")
		repoFile = repo.NewFile()
	}

	log.Printf("creating new chart repostiroy...")
	r, err := repo.NewChartRepository(&entry, getter.All(settings))
	if err != nil {
		return err
	}

	log.Printf("downloading index file for repository...")
	_, err = r.DownloadIndexFile()
	if err != nil {
		return err
	}

	log.Printf("adding repository to repo file...")
	repoFile.Update(&entry)

	log.Printf("writing repo file...")
	repoFile.WriteFile(settings.RepositoryConfig, 0644)
	if err != nil {
		return err
	}

	log.Printf("repository '%s' initialized", repoUrl)

	return nil
}

func createActionConfig(targetNamespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), targetNamespace, os.Getenv("HELM_DRIVER"), log.Printf)

	return actionConfig, settings, err
}

func getChart(chartPathOption action.ChartPathOptions, chartName string, settings *cli.EnvSettings) (*chart.Chart, error) {
	chartPath, err := chartPathOption.LocateChart(chartName, settings)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	return chart, nil
}
