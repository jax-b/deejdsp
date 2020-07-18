package main

import (
	"fmt"
	"time"

	"github.com/jax-b/deej"
	"github.com/jax-b/deejdsp"
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
	d, err := deej.NewDeej(logger, true)
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

	// Load Config
	cfgDSP, err := deejdsp.NewDSPConfig(modlogger)
	cfgDSP.Load()

	if cfgDSP.StartupDelay > 0 {
		modlogger.Infof("Sleeping for controller startup: %s milliseconds", cfgDSP.StartupDelay)
		time.Sleep(time.Duration(cfgDSP.StartupDelay) * time.Millisecond)
	}

	//Set up all modules
	// serSD, err := deejdsp.NewSerialSD(serial, modlogger)
	serTSC, err := deejdsp.NewSerialTCA(serial, modlogger)
	serDSP, err := deejdsp.NewSerialDSP(serial, modlogger)
	if cfgDSP.CommandDelay > 0 {
		time := time.Duration(cfgDSP.CommandDelay) * time.Millisecond
		// serSD.SetTimeDelay(time)
		serTSC.SetTimeDelay(time)
		serDSP.SetTimeDelay(time)
	}

	//Initalise the Displays
	for i := 0; i <= cfgDSP.DisplayMapping.Length(); i++ {
		value, _ := cfgDSP.DisplayMapping.Get(i)

		serTSC.SelectPort(uint8(i))

		if len(value) > 0 {
			serDSP.SetImage(string(value[0]))
			serDSP.DisplayOn()
			modlogger.Named("Display").Infof("%d: %q", i, string(value[0]))
		} else {
			serDSP.DisplayOff()
		}
	}

	// Detect Config Reload
	go func() {
		configReloadedChannel := d.SubscribeToChanges()

		const stopDelay = 50 * time.Millisecond

		for {
			select {
			case <-configReloadedChannel:
				serial.Pause()

				cfgDSP.Load()

				// let the connection close
				<-time.After(stopDelay)

				for i := 0; i <= cfgDSP.DisplayMapping.Length(); i++ {
					value, _ := cfgDSP.DisplayMapping.Get(i)

					serTSC.SelectPort(uint8(i))

					if len(value) > 0 {
						serDSP.DisplayOn()
						serDSP.SetImage(string(value[0]))
						modlogger.Named("Display").Infof("%d: %q", i, string(value[0]))
					} else {
						serDSP.DisplayOff()
					}
				}

				serial.Start()
			}
		}
	}()

	d.Start()

}
