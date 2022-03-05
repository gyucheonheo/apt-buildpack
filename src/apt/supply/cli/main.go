package main

import (
	"os"
	"fmt"
	"os/ioutil"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/apt-buildpack/src/apt/apt"
	"github.com/cloudfoundry/apt-buildpack/src/apt/supply"

	"github.com/cloudfoundry/libbuildpack"
)

func main() {
	logger := libbuildpack.NewLogger(os.Stdout)

	if os.Getenv("CF_STACK") == libbuildpack.CFLINUXFS2 {
		logger.Error("stack : %s is no longer supported by this buildpack", libbuildpack.CFLINUXFS2)
		os.Exit(8)
	}

	if os.Getenv("BP_DEBUG") != "" {
		cmd := exec.Command("find", ".")
		cmd.Dir = "/tmp/cache"
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		logger.Error("Unable to determine buildpack directory: %s", err.Error())
		os.Exit(9)
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		logger.Error("Unable to load buildpack manifest: %s", err.Error())
		os.Exit(10)
	}

	stager := libbuildpack.NewStager(os.Args[1:], logger, manifest)
	if err := stager.CheckBuildpackValid(); err != nil {
		os.Exit(11)
	}

	if err = stager.SetStagingEnvironment(); err != nil {
		logger.Error("Unable to setup environment variables: %s", err.Error())
		os.Exit(13)
	}
	dataString :="--- \npackages: \n  - firefox \n  - libgtk-3-0 \n  - libx11-xcb1 \n  - libdbus-glib-1-2 \n  - libxt6 \n"
    dataBytes := []byte(dataString)

    ioutil.WriteFile("/tmp/app/apt.yml", dataBytes, 0)

	command := &libbuildpack.Command{}
	a := apt.New(command, "tmp/app/apt.yml", "/etc/apt", stager.CacheDir(), filepath.Join(stager.DepDir(), "apt"), logger)
	if err := a.Setup(); err != nil {
		logger.Error("Unable to initialize apt package: %s", err.Error())
		os.Exit(13)
	}

	supplier := supply.New(stager, a, logger)

	if err := supplier.Run(); err != nil {
		logger.Error("Error running supply: %s", err.Error())
		os.Exit(14)
	}

	if err := stager.WriteConfigYml(nil); err != nil {
		logger.Error("Error writing config.yml: %s", err.Error())
		os.Exit(15)
	}

	stager.StagingComplete()
}
