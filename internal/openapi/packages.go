// package api why do I have to put this in every file and then keep them all up to date, etc
package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gitlab.alpinelinux.org/alpine/go/repository"
)

// TODO
//  * Check if org/arch/repo/etc are in known
//  * Beef up path traversal protection

// TODO this is a transitional var, need to move everything to use pkgrepoapi.packagebasedir
var PackageBaseDirectory string

// PkgRepoAPI - a collection packages, repos, versions, etc
type PkgRepoAPI struct {
	Repos map[string]Repo
	// PackageBaseDirectory - the base directory where packages are organized/stored
	PackageBaseDirectory string
}

// NewPkgRepo - called by main function to
func NewPkgRepo(dir string) *PkgRepoAPI {
	p := &PkgRepoAPI{
		Repos: make(map[string]Repo),
	}
	// PackageBaseDirectory - the base directory where packages are organized/stored
	p.PackageBaseDirectory = "file://" + dir
	PackageBaseDirectory = "file://" + dir

	return p
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendRepoError(ctx echo.Context, code int, message string) {
	repoErr := Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, repoErr)
	// return err
	if err != nil {
		log.Warn().Err(err).Msg("failed to return json error")
	}
}

// ListRepos - list repos in an org
func (p *PkgRepoAPI) ListRepos(ctx echo.Context, org, distro, version string) error {
	// log.Debug().Str("org", org).Msg("ListRepos request")
	result, err := listRepos(org, distro, version)
	if err != nil {
		log.Error().Err(err).Msg("failed to listRepos")
	}
	if len(result) == 0 {
		return ctx.JSON(http.StatusInternalServerError, Error{Code: http.StatusInternalServerError, Message: "org does not have any repos"})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateRepo - Add a new package repo
func (p *PkgRepoAPI) CreateRepo(ctx echo.Context, org string) error {
	// We expect a NewRepo object in the request body.
	var newRepo NewRepo
	err := ctx.Bind(&newRepo)
	if err != nil {
		log.Warn().Err(err).Str("org", org).Msg("failed to create repo")
		sendRepoError(ctx, http.StatusBadRequest, "Invalid format for NewRepo")
	}
	// We now have a repo, let's add it to our "database".
	// TODO

	return ctx.JSON(http.StatusCreated, Repo{Name: newRepo.Name})
}

// FindRepoByName - return a repo info from org, distro, version, repo name
func (p *PkgRepoAPI) FindRepoByName(ctx echo.Context, org, distro, version, repo string) error {
	repoInfo := getRepoInfo(org, distro, version, repo)

	return ctx.JSON(http.StatusOK, []Repo{repoInfo})
}

// DeleteRepo - delete a repo TODO
// func (p *PkgRepoAPI) DeleteRepo(ctx echo.Context, org, distro, version, repo string) error {
// 	return ctx.NoContent(http.StatusNoContent)
// }

// GetHealthPing - return health status (for k8s)
func (p *PkgRepoAPI) GetHealthPing(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

// GetHealtReady - return ready status (for k8s)
func (p *PkgRepoAPI) GetHealthReady(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ready")
}

// HeadHealthReady - return ready status (for k8s)
func (p *PkgRepoAPI) HeadHealthReady(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ready")
}

// ListPackagesByRepo - list packages in an org's repo
func (p *PkgRepoAPI) ListPackagesByRepo(ctx echo.Context, org, distro, version, repo, arch string) error {
	pkgs, err := listPackages(org, distro, version, repo, arch)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "failed to get package list"})
	}

	return ctx.JSON(http.StatusOK, pkgs)
}

// CreatePackage - Create a package in a repo and regenerate the index
func (p *PkgRepoAPI) CreatePackage(ctx echo.Context, org, distro, ver, repo, arch string) error {
	file, err := ctx.FormFile("package")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get file from submitted data")
	}
	src, err := file.Open()
	if err != nil {
		log.Warn().Err(err).Msg("failed to open src file")
	}
	defer src.Close()

	pkg, err := repository.ParsePackage(src)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse package from uploaded file")
		return ctx.JSON(http.StatusInternalServerError, errors.New("failed to parse upload"))
	}
	src.Close()

	writeUploadedPkg(file, org, distro, ver, repo, arch)

	// go generateAPKIndex(org, distro, ver, repo, arch)

	return ctx.JSON(http.StatusOK, &Package{Name: pkg.Name, Version: &pkg.Version}) // TODO do we really pkgrel?
}

// CreatePackageIndex - regenerate the index for a repo
func (p *PkgRepoAPI) CreatePackageIndex(ctx echo.Context, org, distro, ver, repo, arch string) error {
	// generateAPKIndex is meant to be run in a goroutine, so we don't actually get any return from the function
	// for now, just blindly return true, but we should tidy this up later
	generateAPKIndex(org, distro, ver, repo, arch)

	status := &GenerateIndex{Status: true}
	return ctx.JSON(http.StatusOK, status)
}

// ListDistros - return a list of the supported distros
func (p *PkgRepoAPI) ListDistros(ctx echo.Context, org string) error {
	distros, err := listDistros(org)
	if err != nil {
		log.Error().Err(err).Msg("failed to listDistros")
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "failed to get a list of distros"})
	}
	return ctx.JSON(http.StatusOK, distros)
}

// ListOrganizations - return a list of the organizations
// TODO this should probably only be available to the server tokens
func (p *PkgRepoAPI) ListOrganizations(ctx echo.Context) error {
	orgs := listOrgs()
	return ctx.JSON(http.StatusOK, orgs)
}

// GetOrganization - get an org
func (p *PkgRepoAPI) GetOrganization(ctx echo.Context, org string) error {
	ret := &Organization{}
	if orgExists(org) {
		// distros, err := listRepos(org, distro)
		ret = &Organization{
			Name:          &org,
			Distributions: nil,
		}
	}

	return ctx.JSON(http.StatusOK, ret)
}

// ListVersions - list of versions in an org's repo
func (p *PkgRepoAPI) ListVersions(ctx echo.Context, org, distro string) error {
	dvs, err := listVersions(org, distro)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "failed to list distro versions"})
	}
	return ctx.JSON(http.StatusOK, dvs)
}

// GetOrgDistro - Return info about a distribution for an org
func (p *PkgRepoAPI) GetOrgDistro(ctx echo.Context, org, distro string) error {
	log.Debug().Str("org", org).Str("distro", distro).Msg("GetOrgDistro")
	return ctx.JSON(http.StatusOK, []Distribution{})
}

// ListArches - list architectures for org/distro/version/repo
func (p *PkgRepoAPI) ListArches(ctx echo.Context, org, distro, version, repo string) error {
	arches, err := listArches(org, distro, version, repo)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "failed to get arch list"})
	}

	return ctx.JSON(http.StatusOK, arches)

}
