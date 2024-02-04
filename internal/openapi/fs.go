// package api provides the api... shocker huh?
package api

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"gitlab.alpinelinux.org/alpine/go/repository"
)

// GetValidTokens - return an array of token strings
// this is exported because it gets called in the auth validator in cmd/api/main.go
func GetValidTokens(org string) []string {
	ctx := context.Background()
	var tokens []string
	configURI := PackageBaseDirectory + "/config/" + org + "/tokens/" // filepath.join condenses the consectutive
	tokenFS := afs.New()
	err := tokenFS.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Msg("GetValidTokens: failed to create NewLocation")
	}

	tokenList, err := tokenFS.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to list the token dir")
	}
	for _, t := range tokenList {
		if t.IsDir() {
			continue
		}
		fd, err := tokenFS.Open(ctx, t)
		if err != nil || fd == nil {
			log.Error().Err(err).Msg("GetValidTokens: failed to create new file from token path")
		}
		var tok = make([]byte, 256)
		c, err := fd.Read(tok)
		if err != nil {
			log.Error().Err(err).Int("read count", c).Msg("failed to read from fd")
		}
		tok = bytes.TrimSpace(tok[0:c])
		if len(tok) > 0 {
			tokens = append(tokens, string(tok))
		}
	}

	return tokens
}

func listOrgs() []Organization {
	var orgs []Organization

	ctx := context.Background()
	configURI := url.JoinUNC(PackageBaseDirectory, "static")
	cfs := afs.New()
	err := cfs.Init(ctx, configURI)
	if err != nil {
		log.Fatal().Err(err).Str("uri", configURI).Msg("listOrgs: failed to init afs")
	}

	orgDirList, err := cfs.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to list orgs")
	}
	for _, o := range orgDirList {
		if o.URL() == url.Normalize(configURI, file.Scheme) {
			continue
		}
		orgName := o.Name()
		orgs = append(orgs, Organization{Name: &orgName})
	}
	return orgs
}

func orgExists(org string) bool {
	orgs := listOrgs()
	for _, o := range orgs {
		if *o.Name == org {
			return true
		}
	}
	return false
}

func listRepos(org, distro, version string) ([]Repo, error) {
	ctx := context.Background()
	configURI := url.JoinUNC(PackageBaseDirectory, "static", org, distro, version)
	cfs := afs.New()
	err := cfs.Init(ctx, configURI)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Str("distro", distro).Msg("failed to list repos")
	}
	result := []Repo{}

	// TODO should we use list instead of walk?
	walkerF := func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if parent == "" {
			// if parent is set that means we've recursed
			result = append(result, Repo{Name: info.Name()})
		}
		return true, nil
	}

	err = cfs.Walk(ctx, configURI, walkerF)
	if err != nil {
		log.Error().Err(err).Msg("listRepos: failed to walk dirs")
	}

	return result, nil
}

func listPackages(org, distro, version, repo, arch string) ([]Package, error) {
	ctx := context.Background()
	configURI := url.JoinUNC(PackageBaseDirectory, "static", org, distro, version, repo, arch)
	cfs := afs.New()
	err := cfs.Init(ctx, configURI)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Str("distro", distro).Msg("failed to init afs")
	}
	result := []Package{}

	walkerF := func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if strings.HasSuffix(info.Name(), ".apk") {
			if parent == "" {
				// if parent is set that means we've recursed
				result = append(result, Package{Name: info.Name()})
			}
		}
		return true, nil
	}

	err = cfs.Walk(ctx, configURI, walkerF)
	if err != nil {
		log.Error().Err(err).Msg("listPackages: failed to walk dirs")
	}

	return result, nil
}

func getRepoInfo(org, distro, version, repo string) Repo {
	ctx := context.Background()
	configURI := url.JoinUNC(PackageBaseDirectory, "config", org, distro, version, repo)
	cfs := afs.New()
	err := cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Str("distro", distro).Msg("getRepoInfo: failed to init afs")
	}

	ex, err := cfs.Exists(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to check existance of repo")
	}
	if ex {
		return Repo{
			Name:        repo,
			Description: nil,
			// Architectures: ,
		}
	}
	return Repo{}
}

func listVersions(org, distro string) ([]RepoVersion, error) {
	result := []RepoVersion{}
	ctx := context.Background()
	// this looks like it's hardcoding the scheme, but it's really just trying to duplicate the logic that afs.List() uses
	configURI := url.Normalize(url.JoinUNC(PackageBaseDirectory, "config", org, distro), file.Scheme)
	cfs := afs.New()
	err := cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Str("distro", distro).Msg("listVersions: failed to init afs")
	}
	vList, err := cfs.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to list repoversions")
	}
	for _, vers := range vList {
		if vers.URL() == configURI || !vers.IsDir() {
			// we can skip the parent dir and files
			continue
		}
		result = append(result, vers.Name())
	}
	return result, nil
}
func listArches(org, distro, version, repo string) ([]Architecture, error) {
	result := []Architecture{}
	ctx := context.Background()
	// this looks like it's hardcoding the scheme, but it's really just trying to duplicate the logic that afs.List() uses
	configURI := url.Normalize(url.JoinUNC(PackageBaseDirectory, "static", org, distro, version, repo), file.Scheme)
	cfs := afs.New()
	err := cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Str("distro", distro).Msg("listArches: failed to init afs")
	}
	aList, err := cfs.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to list arch dir")
	}
	for _, a := range aList {
		if a.URL() == configURI || !a.IsDir() {
			// we can skip the parent dir and files
			continue
		}
		result = append(result, a.Name())
	}
	return result, nil
}

func listDistros(org string) ([]Distribution, error) {
	result := []Distribution{}
	ctx := context.Background()
	// this looks like it's hardcoding the scheme, but it's really just trying to duplicate the logic that afs.List() uses
	configURI := url.Normalize(url.JoinUNC(PackageBaseDirectory, "static", org), file.Scheme)
	cfs := afs.New()
	err := cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Str("org", org).Str("uri", configURI).Msg("failed to list repos")
	}
	dList, err := cfs.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Msg("failed to list repoversions")
	}
	for _, distro := range dList {
		if distro.URL() == configURI || !distro.IsDir() {
			// we can skip the parent dir and files
			continue
		}
		result = append(result, distro.Name())
	}
	return result, nil
}

func writeUploadedPkg(f *multipart.FileHeader, org, distro, version, repo, arch string) {
	log.Debug().Msg("writing uploaded package")
	ctx := context.Background()
	staticURI := url.JoinUNC(PackageBaseDirectory, "static", org, distro, version, repo, arch)

	pkgFile, err := f.Open()
	if err != nil {
		log.Error().Err(err).Msg("failed to open pkg file")
	}

	cfs := afs.New()
	err = cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Msg("generateAPKIndex: failed to init cfs")
	}
	outFileName := url.JoinUNC(staticURI, f.Filename)
	_ = cfs.Delete(ctx, outFileName) // we don't care if delete fails as the file probably doesn't even exist
	outFile, err := cfs.NewWriter(ctx, outFileName, 0644)
	if err != nil {
		log.Error().Err(err).Msg("failed to create outfile")
	}

	c, err := io.Copy(outFile, pkgFile)
	if err != nil || c == 0 {
		log.Error().Err(err).Int64("count", c).Msg("failed to copy uploaded pkg file to outFile")
	}
}

// generateAPKIndex - (re)generate the APKINDEX file
// this runs in the background because it can take quite a while
// regenerate the APKINDEX
// TODO make sure we don't run this unnecessarily
func generateAPKIndex(org, distro, version, repo, arch string) {
	log.Debug().Msg("generating APK index")
	var apki repository.ApkIndex
	apki.Description = "atlascloud main edge"

	ctx := context.Background()
	configURI := url.JoinUNC(PackageBaseDirectory, "config", org, distro)
	staticURI := url.JoinUNC(PackageBaseDirectory, "static", org, distro, version, repo, arch)
	cfs := afs.New()
	err := cfs.Init(ctx, PackageBaseDirectory)
	if err != nil {
		log.Fatal().Err(err).Msg("generateAPKIndex: failed to init cfs")
	}

	walkerF := func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		// log.Debug().Str("baseurl", baseURL).Str("parent", parent).Msg("generateAPKIndex: walker")
		if strings.HasSuffix(info.Name(), ".apk") {
			if parent == "" {
				// if parent is set that means we've recursed
				pkg, err := repository.ParsePackage(reader)
				if err != nil {
					log.Error().Err(err).Msg("generateAPKIndex: failed to parse package")
				}
				apki.Packages = append(apki.Packages, pkg)
			}
		}
		return true, nil
	}

	err = cfs.Walk(ctx, staticURI, walkerF)
	if err != nil {
		log.Error().Err(err).Str("URI", staticURI).Msg("failed to walk")
	}

	distroDirContents, err := cfs.List(ctx, configURI)
	if err != nil {
		log.Error().Err(err).Str("uri", configURI).Msg("failed to list rsa key")
	}

	// the afs matcher thing doesn't seem to work, so we have to find it the hard way
	var keyFd io.ReadCloser
	var keyName string
	for _, f := range distroDirContents {
		if strings.HasSuffix(f.Name(), ".rsa") {
			keyFd, err = cfs.Open(ctx, f)
			if err != nil {
				log.Error().Err(err).Str("uri", configURI).Msg("failed to open rsa key")
			}
			keyName = f.Name()
		}
	}

	keyData, err := io.ReadAll(keyFd)
	if err != nil {
		log.Error().Err(err).Str("uri", configURI).Msg("failed to read rsa key")
	}

	der, _ := pem.Decode(keyData)

	key, err := x509.ParsePKCS1PrivateKey(der.Bytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse key from der")
	}

	archive, err := repository.ArchiveFromIndex(&apki)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate archive from index")
	}
	signedarchive, err := repository.SignArchive(archive, key, keyName+".pub")
	if err != nil {
		log.Error().Err(err).Msg("failed to sign archive")
	}
	sabytes, err := io.ReadAll(signedarchive)
	if err != nil {
		log.Error().Err(err).Msg("failed to read signed archive bytes")
	}

	outFilePath := url.JoinUNC(staticURI, "APKINDEX.tar.gz")
	_ = cfs.Delete(ctx, outFilePath) // we don't care if delete fails as the file may not even exist
	outFile, err := cfs.NewWriter(ctx, outFilePath, 0644)
	if err != nil {
		log.Error().Err(err).Msg("failed to create outfile")
	}
	c, err := outFile.Write(sabytes)
	if err != nil || c == 0 {
		log.Error().Err(err).Int("count", c).Msg("failed to write signed archive")
	}
}
