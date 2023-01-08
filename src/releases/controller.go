package releases

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zsfarkas/chartinstaller/src/generic"
)

type Controller struct {
	ChartMuseumUri  string
	TargetNamespace string
}

func NewController() *Controller {
	targetNamespace, present := os.LookupEnv("TARGET_NAMESPACE")
	if !present {
		targetNamespace = "default"
	}

	repoUrl := os.Getenv("CHART_MUSEUM_URI")

	err := initRepo(repoUrl, targetNamespace)
	if err != nil {
		log.Printf("could not initialize repository %s; cause: %s", repoUrl, err)
	}

	return &Controller{
		ChartMuseumUri:  repoUrl,
		TargetNamespace: targetNamespace,
	}
}

// @Summary      List the config of the service
// @Description  It lists all the properties, which are configured for this service.
// @Tags         releases
// @Accept       json
// @Produce      json
// @Success      200  {object}  Controller
// @Failure      400  {object}  generic.HTTPError
// @Failure      500  {object}  generic.HTTPError
// @Router       /releases/config [get]
func (c *Controller) GetConfig(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c)
}

// @Summary      List all of the chart releases names
// @Description  It lists all releases names, which were previously installed.
// @Tags         releases
// @Accept       json
// @Produce      json
// @Success      200  {array}  string
// @Failure      400  {object}  generic.HTTPError
// @Failure      500  {object}  generic.HTTPError
// @Router       /releases [get]
func (c *Controller) ListReleases(ctx *gin.Context) {
	releases, err := getReleaseNames(c.TargetNamespace)

	if err != nil {
		generic.NewError(ctx, 400, err)
		return
	}

	ctx.JSON(http.StatusOK, releases)
}

// @Summary      Get the status of one chart release
// @Description  It gets the status information of one chart release, which was previously installed.
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        name path      string  true  "Name"
// @Success      200  {object}  ReleaseStatus
// @Failure      400  {object}  generic.HTTPError
// @Failure      500  {object}  generic.HTTPError
// @Router       /releases/{name} [get]
func (c *Controller) StatusRelease(ctx *gin.Context) {
	name := ctx.Param("name")
	if len(name) == 0 {
		generic.NewError(ctx, 400, errors.New("name param must be provided"))
		return
	}

	releaseStatus, err := getReleaseStatus(name, c.TargetNamespace)
	if err != nil {
		generic.NewError(ctx, 400, err)
		return
	}

	ctx.JSON(http.StatusOK, releaseStatus)
}

// @Summary      Uninstall a chart release
// @Description  It uninstalls the given chart release, which was previously installed.
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        name path      string  true  "Name"
// @Success      200  {object}  ReleaseStatus
// @Failure      400  {object}  generic.HTTPError
// @Failure      500  {object}  generic.HTTPError
// @Router       /releases/{name} [delete]
func (c *Controller) UninstallRelease(ctx *gin.Context) {
	name := ctx.Param("name")
	if len(name) == 0 {
		generic.NewError(ctx, 400, errors.New("name param must be provided"))
		return
	}

	releaseStatus, err := uninstallRelease(name, c.TargetNamespace)
	if err != nil {
		generic.NewError(ctx, 400, err)
		return
	}

	ctx.JSON(http.StatusOK, releaseStatus)
}

// @Summary      Install or upgrade one chart release
// @Description  It installs one chart with the provided release name and values. If the release is already installed, then it upgrades the release with the provided values.
// @Tags         releases
// @Accept       json
// @Produce      json
// @Param        name path      string  true  "Name"
// @Param        releaseRequest body   ReleaseRequest true "{"chart": "grafana", "values": {}}"
// @Success      200  {object}  ReleaseStatus
// @Failure      400  {object}  generic.HTTPError
// @Failure      500  {object}  generic.HTTPError
// @Router       /releases/{name} [put]
func (c *Controller) InstallOrUpgradeRelease(ctx *gin.Context) {
	name := ctx.Param("name")
	if len(name) == 0 {
		generic.NewError(ctx, 400, errors.New("name param must be provided"))
		return
	}

	var releaseRequest ReleaseRequest
	if err := ctx.ShouldBindJSON(&releaseRequest); err != nil {
		generic.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := releaseRequest.Validation(); err != nil {
		generic.NewError(ctx, http.StatusBadRequest, err)
		return
	}

	releaseStatus, err := installOrUpgradeRelease(name, c.ChartMuseumUri, releaseRequest.Chart, releaseRequest.Values, c.TargetNamespace)
	if err != nil {
		generic.NewError(ctx, 400, err)
		return
	}

	ctx.JSON(http.StatusOK, releaseStatus)
}
