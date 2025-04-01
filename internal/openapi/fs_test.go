package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidTokens(t *testing.T) {
	// Setup test directory structure
	tmpDir, err := os.MkdirTemp("", "test-tokens-*")
	if err != nil {
		t.Fatal("failed to create testGetValidTokens tmpDir", err)
	}
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal("failed to remove testGetValidTokens tmpDir", err)
		}
	}()

	// Set the package base directory for testing
	originalDir := PackageBaseDirectory
	PackageBaseDirectory = "file://" + tmpDir
	defer func() { PackageBaseDirectory = originalDir }()

	// Create test organization directory structure
	org := "testorg"
	tokenDir := tmpDir + "/config/" + org + "/tokens"
	err = os.MkdirAll(tokenDir, 0755)
	if err != nil {
		t.Fatal("failed to create testGetValidTokens path", err)
	}

	// Create test token files
	testTokens := []string{"token1", "token2"}
	for i, token := range testTokens {
		err = os.WriteFile(fmt.Sprintf("%s/%d", tokenDir, i), []byte(token), 0644)
		if err != nil {
			t.Fatal("failed to create testGetValidTokens file", err)
		}
	}

	// Test GetValidTokens
	tokens := GetValidTokens(org)
	assert.Len(t, tokens, 2)
	assert.Contains(t, tokens, "token1")
	assert.Contains(t, tokens, "token2")
}

func TestListDistros(t *testing.T) {
	// Setup test directory structure
	tmpDir, err := os.MkdirTemp("", "test-distros-*")
	if err != nil {
		t.Fatal("failed to create testListDistros tmpDir", err)
	}
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal("failed to remove testListDistros tmpDir", err)
		}
	}()

	// Set the package base directory for testing
	originalDir := PackageBaseDirectory
	PackageBaseDirectory = "file://" + tmpDir
	defer func() { PackageBaseDirectory = originalDir }()

	// Create test organization and distro structure
	org := "testorg"
	staticDir := tmpDir + "/static/" + org
	testDistros := []string{"alpine", "ubuntu"}

	for _, distro := range testDistros {
		err = os.MkdirAll(staticDir+"/"+distro, 0755)
		if err != nil {
			t.Fatal("failed to create testListDistros path", err)
		}
	}

	// Test ListDistros
	distros, err := listDistros(org)
	assert.NoError(t, err)
	assert.Len(t, distros, 2)

	assert.Contains(t, distros, "alpine")
	assert.Contains(t, distros, "ubuntu")
}

func TestListArches(t *testing.T) {
	// Setup test directory structure
	tmpDir, err := os.MkdirTemp("", "test-arches-*")
	if err != nil {
		t.Fatal("failed to create testListArches tmpDir", err)
	}
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal("failed to remove testListArches tmpDir", err)
		}
	}()

	// Set the package base directory for testing
	originalDir := PackageBaseDirectory
	PackageBaseDirectory = "file://" + tmpDir
	defer func() { PackageBaseDirectory = originalDir }()

	// Create test directory structure
	org := "testorg"
	distro := "alpine"
	version := "3.14"
	repo := "main"
	path := tmpDir + "/static/" + org + "/" + distro + "/" + version + "/" + repo

	testArches := []string{"x86_64", "aarch64"}
	for _, arch := range testArches {
		err = os.MkdirAll(path+"/"+arch, 0755)
		if err != nil {
			t.Fatal("failed to create testListArches path", err)
		}
	}

	// Test ListArches
	arches, err := listArches(org, distro, version, repo)
	assert.NoError(t, err)
	assert.Len(t, arches, 2)
	assert.Contains(t, arches, "x86_64")
	assert.Contains(t, arches, "aarch64")
}

func TestListOrgs(t *testing.T) {
	// Setup test directory structure
	tmpDir, err := os.MkdirTemp("", "test-orgs-*")
	if err != nil {
		t.Fatal("failed to create testListOrgs tmpDir", err)
	}
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Fatal("failed to remove testListOrgs tmpDir", err)
		}
	}()

	// Set the package base directory for testing
	originalDir := PackageBaseDirectory
	PackageBaseDirectory = "file://" + tmpDir
	defer func() { PackageBaseDirectory = originalDir }()

	// Create test organizations
	testOrgs := []string{"org1", "org2"}
	for _, org := range testOrgs {
		err = os.MkdirAll(tmpDir+"/static/"+org, 0755)
		if err != nil {
			t.Fatal("failed to create testListOrgs path", err)
		}
	}

	// Test ListOrgs
	orgs := listOrgs()
	assert.Len(t, orgs, 2)

	orgNames := make([]string, len(orgs))
	for i, o := range orgs {
		orgNames[i] = *o.Name
	}
	assert.Contains(t, orgNames, "org1")
	assert.Contains(t, orgNames, "org2")
}
