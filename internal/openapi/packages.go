package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// TODO
//  * Check if org/arch/repo/etc are in known
//  * Beef up path traversal protection

// PackageBaseDirectory - the base directory where packages are organized/stored
var PackageBaseDirectory = "/srv/packages"

// PkgRepoAPI - a collection packages, repos, versions, etc
type PkgRepoAPI struct {
	Repos          map[string]Repo
	DistroVersions []*DistroVersion
}

// NewPkgRepo - called by main function to
func NewPkgRepo() *PkgRepoAPI {
	p := &PkgRepoAPI{
		Repos: make(map[string]Repo),
	}

	return p
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendRepoError(ctx echo.Context, code int, message string) error {
	repoErr := Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, repoErr)
	return err
}

// ListRepos - list repos in an org
func (p *PkgRepoAPI) ListRepos(ctx echo.Context, org string) error {
	var result []Repo

	repos, err := ioutil.ReadDir(filepath.Join(PackageBaseDirectory, filepath.Clean(org), "alpine"))
	if err != nil {
		log.Warn().Err(err).Msg("failed to readdir org")
		sendRepoError(ctx, http.StatusInternalServerError, fmt.Sprintf("ListRepos: %s", err))
	}

	for _, r := range repos {
		log.Trace().Interface("repos -> r", r).Msg("repos loop")
		if r.IsDir() {
			nents, err := ioutil.ReadDir(filepath.Join(PackageBaseDirectory, filepath.Clean(org), "alpine", r.Name()))
			if err != nil {
				log.Warn().Err(err).Msg("failed to readdir r.Name")
				sendRepoError(ctx, http.StatusInternalServerError, fmt.Sprintf("ListRepos: %s", err))
			}
			for _, ne := range nents {
				// fmt.Println(e, ne)
				r := Repo{
					Name:    ne.Name(),
					Version: r.Name(),
					// Slug:    slug.Make(fmt.Sprintf("%s %s", e.Name(), ne.Name())),
				}

				// p.Repos[r.Slug] = r

				result = append(result, r)
			}
		}
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateRepo - Add a new package repo
func (p *PkgRepoAPI) CreateRepo(ctx echo.Context, org string) error {
	// We expect a NewRepo object in the request body.
	var newRepo NewRepo
	err := ctx.Bind(&newRepo)
	if err != nil {
		return sendRepoError(ctx, http.StatusBadRequest, "Invalid format for NewRepo")
	}
	// We now have a repo, let's add it to our "database".

	// We handle repos, not NewRepos, which have an additional ID field
	var repo Repo
	repo.Name = newRepo.Name

	// Now, we have to return the NewRepo
	err = ctx.JSON(http.StatusCreated, repo)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	// Return no error. This refers to the handler. Even if we return an HTTP
	// error, but everything else is working properly, tell Echo that we serviced
	// the error. We should only return errors from Echo handlers if the actual
	// servicing of the error on the infrastructure level failed. Returning an
	// HTTP/400 or HTTP/500 from here means Echo/HTTP are still working, so
	// return nil.
	return nil
}

// FindRepoByName - return a pkgrepo from an
func (p *PkgRepoAPI) FindRepoByName(ctx echo.Context, org, repo string) error {
	var repoInfo []Repo

	repos, err := filepath.Glob(filepath.Join(PackageBaseDirectory, filepath.Clean(org), "alpine", "*", repo))
	if err != nil {
		log.Warn().Err(err).Msg("failed to glob repo")
		sendRepoError(ctx, http.StatusInternalServerError, fmt.Sprintf("ListRepos: %s", err))
	}
	for _, ne := range repos {
		// TODO seems like there should be a more foolproof way, but this works for now
		spl := strings.Split(ne, "/")
		r := Repo{
			Name:    repo,
			Version: spl[len(spl)-2 : len(spl)-1][0],
		}

		// p.Repos[r.Slug] = r

		repoInfo = append(repoInfo, r)
	}

	return ctx.JSON(http.StatusOK, repoInfo)
}

// DeleteRepo - delete a repo
func (p *PkgRepoAPI) DeleteRepo(ctx echo.Context, slug string) error {
	return ctx.NoContent(http.StatusNoContent)
}

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

// ListDistroVersions - return a list of distroversions
func (p *PkgRepoAPI) ListDistroVersions(ctx echo.Context, distro string) error {
	var d []DistroVersion

	ents, err := ioutil.ReadDir(filepath.Join(PackageBaseDirectory, "atlascloud", filepath.Clean(distro)))
	if err != nil {
		sendRepoError(ctx, http.StatusInternalServerError, fmt.Sprintf("ListDistroVersions: %s", err))
	}

	for _, e := range ents {
		log.Debug().Interface("e", e).Msg("range distros")
		if e.IsDir() {
			d = append(d, DistroVersion(e.Name()))
		}
	}

	return ctx.JSON(http.StatusOK, d)
}

// func slugToPath(slug string) (string, error) {
// 	// TODO see if path actually exists, etc
// 	s := strings.SplitN(slug, "-", 2)
// 	v := s[0]
// 	r := s[1]

// 	path := fmt.Sprintf("/%s/%s", v, r)

// 	return path, nil
// }

func parseAPKFilename(filename string) (string, string, string) {
	firstPart := filename[0:strings.LastIndex(filename, "-")]
	name := firstPart[:strings.LastIndex(firstPart, "-")]
	version := firstPart[strings.LastIndex(firstPart, "-")+1:]
	release := strings.TrimSuffix(filename[strings.LastIndex(filename, "-")+2:], ".apk")

	return name, version, release
}

// ListPackagesByRepo - asdf
func (p *PkgRepoAPI) ListPackagesByRepo(ctx echo.Context, org, repo, ver string) error {
	// log.Debug().Str("slug", slug).Msg("ListPackagesByRepo")
	var pkgs []Package

	// pkgPath, err := slugToPath(slug)
	// pkgPath := fmt.Sprintf("/%s/%s", ver, repo)

	// if err != nil {
	// 	sendRepoError(ctx, http.StatusInternalServerError, fmt.Sprintf("Invalid slug: %s", err))
	// }

	ents, err := ioutil.ReadDir(path.Join(PackageBaseDirectory, filepath.Clean(org), "alpine", filepath.Clean(ver), filepath.Clean(repo), "/x86_64"))
	if err != nil {
		log.Error().Err(err).Msg("ListPackagesByRepo: ReadDir")
		sendRepoError(ctx, http.StatusInternalServerError, "ListPackagesByRepo: ReadDir error")
	}

	for _, e := range ents {
		// fmt.Println(e)
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".apk") {
			// some example file names to deal with
			// abuild-3.7.0_rc1-r0.apk
			// shared-mime-info-lang-1.15-r0.apk
			filename := e.Name()

			name, version, release := parseAPKFilename(filename)

			pkgs = append(pkgs, Package{
				Name:    name,
				Version: &version,
				Release: &release,
			})
		}
	}

	return ctx.JSON(http.StatusOK, pkgs)
}

// CreatePackage - Create a package in a repo and regenerate the index
func (p *PkgRepoAPI) CreatePackage(ctx echo.Context, org, repo, ver string) error {
	arch := filepath.Clean(ctx.FormValue("architecture"))
	file, err := ctx.FormFile("package")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get file from submitted data")
	}
	pkgDir := filepath.Join(PackageBaseDirectory, filepath.Clean(org), "alpine", filepath.Clean(ver), filepath.Clean(repo), arch)
	// log.Debug().
	// 	Interface("ctx", ctx).
	// 	Interface("arch", arch).
	// 	Interface("file", file).
	// 	Str("org", org).
	// 	Str("repo", repo).
	// 	Str("ver", ver).
	// 	Msg("echo context")

	src, err := file.Open()
	if err != nil {
		log.Warn().Err(err).Msg("failed to open src file")
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(filepath.Join(pkgDir, filepath.Clean(file.Filename)))
	if err != nil {
		log.Warn().Err(err).Msg("failed to create dst file")
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.Warn().Err(err).Msg("failed to copy file from src to dst")
	}

	// regenerate the APKINDEX
	// TODO make sure we don't run this unnecessarily
	go func() {
		// generate the index and sign it all in the background so the user isn't waiting around
		// for a possibly long running task
		apks, err := filepath.Glob(filepath.Join(pkgDir, "*.apk"))
		if err != nil {
			log.Warn().Err(err).Msg("failed to glob apks")
		}
		// c := exec.Command("/sbin/apk", "index", "-o", "APKINDEX.new.tar.gz", "-x", "APKINDEX.tar.gz", "-d", "atlascloud main edge")
		args := []string{"index", "--no-warnings", "-o", "APKINDEX.new.tar.gz", "-d", "atlascloud main edge"}
		args = append(args, apks...)
		c := exec.Command("/sbin/apk", args...)

		c.Dir = pkgDir
		stderr, err := c.StderrPipe()
		if err != nil {
			log.Error().Err(err).Msg("error setting up stderr for apk index command")
		}
		stdout, err := c.StdoutPipe()
		if err != nil {
			log.Error().Err(err).Msg("error setting up stdout for apk index command")
		}

		err = c.Start()
		if err != nil {
			log.Error().Str("command", c.String()).Err(err).Msg("failed to run apk index")
			return
		}

		se, err := ioutil.ReadAll(stderr)
		if err != nil {
			log.Error().Err(err).Msg("failed to read stderr")
		}
		log.Warn().Bytes("stderr", se).Msg("apk index stderr")
		so, err := ioutil.ReadAll(stdout)
		if err != nil {
			log.Error().Err(err).Msg("failed to read stdout")
		}
		log.Warn().Bytes("stdout", so).Msg("apk index stdout")

		key_file := os.Getenv("KEY_FILE")
		c = exec.Command("/usr/bin/abuild-sign", "-k", key_file, "APKINDEX.new.tar.gz")
		c.Dir = pkgDir
		stderr, err = c.StderrPipe()
		if err != nil {
			log.Error().Err(err).Msg("error setting up stderr for abuild-sign command")
		}
		stdout, err = c.StdoutPipe()
		if err != nil {
			log.Error().Err(err).Msg("error setting up stdout for abuild-sign command")
		}

		err = c.Start()
		if err != nil {
			log.Error().Str("command", c.String()).Err(err).Msg("failed to run abuild-sign")
			return
		}
		se, err = ioutil.ReadAll(stderr)
		if err != nil {
			log.Error().Err(err).Msg("failed to read stderr")
		}
		log.Warn().Bytes("stderr", se).Msg("abuild-sign stderr")
		so, err = ioutil.ReadAll(stdout)
		if err != nil {
			log.Error().Err(err).Msg("failed to read stdout")
		}
		log.Warn().Bytes("stdout", so).Msg("abuild-sign stdout")

		os.Rename(filepath.Join(pkgDir, "APKINDEX.new.tar.gz"), filepath.Join(pkgDir, "APKINDEX.tar.gz"))
	}()

	name, version, release := parseAPKFilename(file.Filename)

	return ctx.JSON(http.StatusOK, &Package{Name: name, Version: &version, Release: &release})
}

// ListDistros - return a list of the supported distros
func (p *PkgRepoAPI) ListDistros(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []Distribution{"alpine"})
}

// ListOrganizations - return a list of the supported distros
func (p *PkgRepoAPI) ListOrganizations(ctx echo.Context) error {
	orgs := listOrgs()
	var ret []Organization
	for _, o := range orgs {
		ret = append(ret, Organization{Name: &o})
	}
	return ctx.JSON(http.StatusOK, ret)
}

// GetOrganization - get an org
func (p *PkgRepoAPI) GetOrganization(ctx echo.Context, org string) error {
	var orgName = "atlascloud"
	return ctx.JSON(http.StatusOK, &Organization{Name: &orgName})
}

// ListVersions - list of versions in an org's repo
func (p *PkgRepoAPI) ListVersions(ctx echo.Context, org, repo string) error {
	return ctx.JSON(http.StatusOK, DistroVersion("3.12"))
}
