package api

import (
	"path/filepath"
	"strings"

	"github.com/pkg/xattr"
	"github.com/rs/zerolog/log"
)

// GetValidTokens - return an array of token strings
func GetValidTokens(org, repo string) []string {
	// log.Debug().Interface("ctx", ctx).Msg("GetValidTokens: context")
	var tokens []string
	orgDir := filepath.Join(PackageBaseDirectory, org)

	// xattrs := make([]string, 1024)

	xattrs, err := xattr.List(orgDir)
	if err != nil {
		log.Warn().Err(err).Msg("failed to listxattrs on org")
	}
	log.Warn().Strs("xattrs", xattrs).Str("path", orgDir).Msg("org level xattrs")

	for _, xa := range xattrs {
		if strings.HasPrefix(xa, "user.token") {
			t, err := xattr.Get(orgDir, xa)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get xattr")
				continue
			}
			tokens = append(tokens, string(t))
		}
	}
	log.Info().Strs("tokens", tokens).Msg("org tokens")

	return tokens
}

func listOrgs() []string {
	// orgDir := filepath.Join(PackageBaseDirectory, org)

	return []string{"atlascloud"}
}
