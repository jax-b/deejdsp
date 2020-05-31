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

	serSD, err := deejdsp.NewSerialSD(serial, modlogger)
	serTSD, err := deejdsp.NewSerialTSC(serial, modlogger)
	serDSP, err := deejdsp.NewSerialDSP(serial, modlogger)

	// Display Image Test
	serTSD.SelectPort(0)
	serDSP.SetImage("1.B")
	serTSD.SelectPort(1)
	serDSP.SetImage("2.B")
	serTSD.SelectPort(2)
	serDSP.SetImage("3.B")
	serTSD.SelectPort(3)
	serDSP.SetImage("4.B")
	serTSD.SelectPort(4)
	serDSP.SetImage("5.B")
	serTSD.SelectPort(5)
	serDSP.SetImage("THECHILD.B")

	// Blink Test
	serTSD.SelectPort(0)
	serDSP.DisplayOff()
	serTSD.SelectPort(1)
	serDSP.DisplayOff()
	serTSD.SelectPort(2)
	serDSP.DisplayOff()
	serTSD.SelectPort(3)
	serDSP.DisplayOff()
	serTSD.SelectPort(4)
	serDSP.DisplayOff()
	serTSD.SelectPort(5)
	serDSP.DisplayOff()

	time.Sleep(500 * time.Millisecond)

	serTSD.SelectPort(0)
	serDSP.DisplayOn()
	serTSD.SelectPort(1)
	serDSP.DisplayOn()
	serTSD.SelectPort(2)
	serDSP.DisplayOn()
	serTSD.SelectPort(3)
	serDSP.DisplayOn()
	serTSD.SelectPort(4)
	serDSP.DisplayOn()
	serTSD.SelectPort(5)
	serDSP.DisplayOn()

	time.Sleep(500 * time.Millisecond)

	serTSD.SelectPort(0)
	serDSP.DisplayOff()
	serTSD.SelectPort(1)
	serDSP.DisplayOff()
	serTSD.SelectPort(2)
	serDSP.DisplayOff()
	serTSD.SelectPort(3)
	serDSP.DisplayOff()
	serTSD.SelectPort(4)
	serDSP.DisplayOff()
	serTSD.SelectPort(5)
	serDSP.DisplayOff()

	time.Sleep(500 * time.Millisecond)

	serTSD.SelectPort(0)
	serDSP.DisplayOn()
	serTSD.SelectPort(1)
	serDSP.DisplayOn()
	serTSD.SelectPort(2)
	serDSP.DisplayOn()
	serTSD.SelectPort(3)
	serDSP.DisplayOn()
	serTSD.SelectPort(4)
	serDSP.DisplayOn()
	serTSD.SelectPort(5)
	serDSP.DisplayOn()

	// List Dir Demo
	test, _ := serSD.ListDir()
	fmt.Print(test)

	// File Send Demo
	// serSD.SendFile(".\\config.yaml", "config")

	// test, _ = serSD.ListDir()
	// fmt.Print(test)

	// File Delete Demo
	// serSD.Delete("config")

	// test, _ = serSD.ListDir()
	// fmt.Print(test)

	d.Start()
}
