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
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog/log"
	"gitlab.alpinelinux.org/alpine/go/apkbuild"
)

func getApkNameFromApkBuild(apkBuildPath string) (string, error) {
	// get info from the APKBUILD to check against what the API says already exists
	fp, err := os.Open(apkBuildPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to open apkbuild file")
	}
	abf := apkbuild.ApkbuildFile{
		PackageName: path.Base(path.Dir(apkBuildPath)),
		Content:     fp,
	}
	parsed, err := apkbuild.Parse(abf, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse apkbuild file")
	}

	apkFilename := fmt.Sprintf("%s-%s-r%s.apk", parsed.Pkgname, parsed.Pkgver, parsed.Pkgrel)
	// log.Debug().Str("apkFilename", apkFilename).Msg("")

	return apkFilename, nil
}
