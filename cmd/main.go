package main

import (
	"fmt"

	"github.com/jax-b/deej"
	deejdsp "github.com/jax-b/deejDSP"
)

var (
	gitCommit  string
	versionTag string
	buildType  string
)

func main() {
	logger, err := deej.NewLogger(buildType)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}

	named := logger.Named("main")
	named.Debug("Created logger")

	named.Infow("Version info",
		"gitCommit", gitCommit,
		"versionTag", versionTag,
		"buildType", buildType)

	// create the deej instance
	d, err := deej.NewDeej(logger)
	if err != nil {
		named.Fatalw("Failed to create deej object", "error", err)
	}

	// set its version info for the tray to show
	if buildType != "" && (versionTag != "" || gitCommit != "") {
		identifier := gitCommit
		if versionTag != "" {
			identifier = versionTag
		}

		versionString := fmt.Sprintf("Version %s-%s", buildType, identifier)
		d.SetVersion(versionString)
	}

	// onwards, to glory
	if err = d.Initialize(); err != nil {
		named.Fatalw("Failed to initialize deej", "error", err)
	}

	modlogger := d.NewNammedLogger("module")
	serial := d.GetSerial()

	deejSD, err := deejdsp.NewSerialSD(serial, modlogger)

	test, _ := deejSD.ListDir()
	fmt.Print(test)

	deejSD.SendFile(".\\config.yaml", "config")

	test, _ = deejSD.ListDir()
	fmt.Print(test)

	deejSD.Delete("config")

	test, _ = deejSD.ListDir()
	fmt.Print(test)

	d.Start()
}
