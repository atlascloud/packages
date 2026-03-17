/*
Copyright © 2024 Iggy Jackson <iggy@iggy.ninja>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"os"

	repoApi "github.com/atlascloud/packages/internal/openapi"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// pkgCmd represents the pkg command
var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Check if a package exists on the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("pkg called")
		// log.Debug().Strs("args", args).Msg("")
		apkb, err := cmd.Flags().GetString("apkbuild")
		if err != nil {
			log.Error().Err(err).Msg("failed to get apkbuild flag")
		}
		// getApkNameFromApkBuild(apkb)
		apkFilename, err := getApkNameFromApkBuild(apkb)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to getApkNameFromAPKBUILD")
		}
		pkgsToken := os.Getenv("PKGS_TOKEN")
		if pkgsToken == "" {
			log.Fatal().Msg("no token found in environ")
		}

		bearerTokenProvider, err := securityprovider.NewSecurityProviderBearerToken(pkgsToken)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init security provider")
		}

		client, err := repoApi.NewClientWithResponses("https://packages.atlascloud.xyz/api", repoApi.WithRequestEditorFn(bearerTokenProvider.Intercept))
		if err != nil {
			log.Error().Err(err).Msg("failed to create client")
		}

		resp, err := client.ListPackagesByRepoWithResponse(context.Background(), "atlascloud", "alpine", "edge", "main", "x86_64")
		if err != nil {
			log.Error().Err(err).Msg("failed to ping")
		}

		// respBody, _ := io.ReadAll(resp.Body)
		// log.Debug().Str("resp", string(resp.Body)).Msg("resp")

		for _, o := range *resp.JSON200 {
			// log.Debug().Interface("package", o.Name).Msg("")
			if o.Name == apkFilename {
				log.Info().Str("package file name", apkFilename).Msg("package already exists on server")
				return
			}
		}
		log.Info().Str("package file name", apkFilename).Msg("package does not exist on server")
	},
}

func init() {
	checkCmd.AddCommand(pkgCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pkgCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pkgCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pkgCmd.Flags().StringP("apkbuild", "a", "", "APKBUILD to check")
}
