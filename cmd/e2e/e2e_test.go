package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/apolo96/gitfresh"
	"github.com/joho/godotenv"
)

var path, _ = os.Getwd()
var cliBinaryPath string = filepath.Join(path, "cli")
var parallelWithPreparation = true

func TestMain(m *testing.M) {
	if err := prepareMain(); err != nil {
		fmt.Println("preparing, error:", err.Error())
		return
	}
	code := m.Run()
	cleanupMain()
	os.Exit(code)
}

func prepareMain() error {
	fmt.Println("= PREPARING")
	/* Compile artifacts (api & cli) */
	var path, _ = os.Getwd()
	var source = strings.Replace(path, "e2e", "", 1)
	var errs []error
	cli := filepath.Join(source, "cli")
	api := filepath.Join(source, "api")
	var wg sync.WaitGroup
	wg.Add(2)
	/* Build API */
	go func() {

		cmd := exec.Command("go", "build", "-o", path, api)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errs = append(errs, errors.New("compiling api "+string(out)))
		}
		fmt.Println("- INFO: SOURCE_API: ", api, "OUT", path)
		wg.Done()
	}()
	/* Build CLI */
	go func() {

		cmd := exec.Command("go", "build", "-ldflags", "-X 'main.devMode=on'", "-o", path, cli)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errs = append(errs, errors.New("compaling cli "+string(out)))
		}
		fmt.Println("- INFO: SOURCE_API: ", cli, "OUT", path)
		wg.Done()
	}()
	wg.Wait()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	if err := os.Chmod(cli, 0755); err != nil {
		return errors.New("allow exec os permissions to cli")
	}
	if err := os.Chmod(api, 0755); err != nil {
		return errors.New("allow exec os permissions to api")
	}
	return nil
}

func cleanupMain() {
	fmt.Println("= CLEANING")
	var path, _ = os.Getwd()
	/* Delete API Binary */
	fmt.Println("- DELETE API Binary")
	apiFile := filepath.Join(path, "api")
	err := os.Remove(apiFile)
	if err != nil {
		fmt.Printf("error deleting file %s: %v\n", apiFile, err)
		return
	}
	/* Delete CLI Binary */
	fmt.Println("- DELETE CLI Binary")
	cliFile := filepath.Join(path, "cli")
	err = os.Remove(cliFile)
	if err != nil {
		fmt.Printf("error deleting file %s: %v\n", cliFile, err)
		return
	}
	/* Delete App Flatfiles */
	fmt.Println("- DELETE APP Directory")
	dir, _ := os.UserHomeDir()
	if err := os.RemoveAll(filepath.Join(dir, gitfresh.APP_FOLDER)); err != nil {
		fmt.Printf("error deleting dir %s: %v\n", dir, err)
		return
	}
}

func TestVersionCLICommand(t *testing.T) {
	if parallelWithPreparation {
		t.Skip("Atomic test not supported")
		t.Parallel()
		// ToDo: Prepare for Atomic Test
	}
	wantErr := false
	expected := "gitfresh version"
	args := []string{"version"}
	cmd := exec.Command(cliBinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if (err != nil) != wantErr {
		t.Fatalf("executing command failed with error: %v , output %s", err, output)
	}
	if !strings.Contains(string(output), expected) {
		t.Errorf("want %q, but got %q", expected, output)
	}
}

func TestConfigCommandFlags(t *testing.T) {
	/* Arrange */
	if parallelWithPreparation {
		t.Skip("Atomic test not supported")
		t.Parallel()
		// ToDo: Prepare for Atomic Test
	}
	err := godotenv.Load()
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	tGitServerToken := os.Getenv("TEST_GIT_SERVER_TOKEN")
	tTunnelToken := os.Getenv("TEST_TUNNEL_TOKEN")
	tTunnelDomain := os.Getenv("TEST_TUNNEL_DOMAIN")
	tGitWorkDir := os.Getenv("TEST_GIT_WORKDIR")
	if tGitServerToken == "" || tTunnelToken == "" || tTunnelDomain == "" || tGitWorkDir == "" {
		t.Fatal("environment vars is not set")
	}
	args := []string{
		"config",
		fmt.Sprintf("-GitServerToken=%s", tGitServerToken),
		fmt.Sprintf("-GitWorkDir=%s", tGitWorkDir),
		fmt.Sprintf("-TunnelDomain=%s", tTunnelDomain),
		fmt.Sprintf("-TunnelToken=%s", tTunnelToken),
	}
	expected := "Config successfully created"
	wantErr := false
	/* Act */
	cmd := exec.Command(cliBinaryPath, args...)
	output, err := cmd.CombinedOutput()
	/* Assert */
	if (err != nil) != wantErr {
		t.Fatalf("executing command failed with error: %v , output %s", err, output)
	}
	if !strings.Contains(string(output), expected) {
		t.Errorf("want %q, but got %q", expected, output)
	}
}

func TestInitCommand(t *testing.T) {
	if parallelWithPreparation {
		t.Skip("Atomic test not supported")
		t.Parallel()
		// ToDo: Prepare for Atomic Test
	}
	wantErr := false
	args := []string{"init"}
	expected := "Repositories to Refresh:"
	cmd := exec.Command(cliBinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if (err != nil) != wantErr {
		t.Fatalf("executing command failed with error: %v , output %s", err, output)
	}
	if !strings.Contains(string(output), expected) {
		t.Errorf("want %q, but got %q", expected, output)
	}
}

func TestStatusCommand(t *testing.T) {
	if parallelWithPreparation {
		t.Skip("Atomic test not supported")
		t.Parallel()
		// ToDo: Prepare for Atomic Test
	}
	wantErr := false
	args := []string{"status"}
	expected := "GitFresh Agent is running"
	cmd := exec.Command(cliBinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if (err != nil) != wantErr {
		t.Fatalf("executing command failed with error: %v , output %s", err, output)
	}
	if !strings.Contains(string(output), expected) {
		t.Errorf("want %q, but got %q", expected, output)
	}
}

func TestSuccessFlow(t *testing.T) {
	t.Parallel()
	parallelWithPreparation = false
	t.Logf("- RUN FLOW WITH parallelWithPreparation: %v", parallelWithPreparation)
	t.Run("version cli command", TestVersionCLICommand)
	t.Run("config cli command", TestConfigCommandFlags)
	t.Run("init cli command", TestInitCommand)
	t.Run("status cli command", TestStatusCommand)
}
