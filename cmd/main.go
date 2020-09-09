package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/jax-b/deej"
	"github.com/jax-b/deejdsp"
	"github.com/sqweek/dialog"
	"go.uber.org/zap"
)

var (
	gitCommit  string
	versionTag string
	buildType  string

	verbose bool

	d          *deej.Deej
	cfgDSP     *deejdsp.DSPCanonicalConfig
	serSD      *deejdsp.SerialSD
	serTCA     *deejdsp.SerialTCA
	serDSP     *deejdsp.SerialDSP
	sessionMap *deej.SessionMap
	sliderMap  *deej.SliderMap
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "show verbose logs (useful for debugging serial)")
	flag.BoolVar(&verbose, "v", false, "shorthand for --verbose")
	flag.Parse()
}

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

	// provide a fair warning if the user's running in verbose mode
	if verbose {
		named.Debug("Verbose flag provided, all log messages will be shown")
	}

	// create the deej instance
	d, err := deej.NewDeej(logger, verbose)
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

	if err = d.Initialize(); err != nil {
		named.Fatalw("Failed to initialize deej", "error", err)
	}

	modlogger := d.NewNammedLogger("module")
	serial := d.GetSerial()
	// Load Config
	cfgDSP, err = deejdsp.NewDSPConfig(modlogger)
	cfgDSP.Load()

	if cfgDSP.StartupDelay > 0 {
		modlogger.Debugf("Sleeping for controller startup: %s milliseconds", cfgDSP.StartupDelay)
		time.Sleep(time.Duration(cfgDSP.StartupDelay) * time.Millisecond)
	}

	//Set up all modules
	serSD, err = deejdsp.NewSerialSD(serial, modlogger)
	serTCA, err = deejdsp.NewSerialTCA(serial, modlogger)
	serDSP, err = deejdsp.NewSerialDSP(serial, modlogger)

	// Tray Menu item: Send Image
	go func() {
		menuItem := <-d.AddMenuItem("Send Image", "Send A image file to the internal SD card")

		time.Sleep(10 * time.Millisecond)

		for {
			<-menuItem.ClickedCh
			filename, err := dialog.File().Filter("ByteImage", "b").Title("SendImage").Load()
			if err != nil {
				break
			}
			PathElements := strings.Split(filename, "\\")
			sdFilename := PathElements[len(PathElements)-1]
			serSD.SendFile(filename, sdFilename)
		}
	}()

	// Tray Menu item: List Files
	go func() {
		menuItem := <-d.AddMenuItem("List Files", "List the files on the sd card")

		time.Sleep(10 * time.Millisecond)

		for {
			<-menuItem.ClickedCh
			files, _ := serSD.ListDir()
			var filesSingle string
			for _, path := range files {
				filesSingle = filesSingle + path + "\n"
			}
			dialog.Message("%s", filesSingle).Title("SD Files").Info()
		}
	}()

	// Tray Menu Item : Turn off displays
	go func() {
		menuItem := <-d.AddMenuItem("Displays Off", "If the config reloads it will turn off all the displays")

		time.Sleep(10 * time.Millisecond)

		for {
			<-menuItem.ClickedCh
			resumeAfter := serial.IsRunning()

			if serial.IsRunning() {
				serial.Pause()
			}

			for i := range cfgDSP.DisplayMapping {
				serTCA.SelectPort(uint8(i))
				serDSP.DisplayOff()
			}

			if resumeAfter {
				serial.Start()
			}
		}
	}()
	_ = serSD

	sessionMap = d.GetSessionMap()
	sliderMap = d.GetSliderMap()

	if cfgDSP.CommandDelay > 0 {
		time := time.Duration(cfgDSP.CommandDelay) * time.Millisecond
		// serSD.SetTimeDelay(time)
		serTCA.SetTimeDelay(time)
		serDSP.SetTimeDelay(time)
	}

	//Initalise the Displays
	loadDSPMapings(modlogger)
	modlogger.Named("Serial").Debug("Flushing")
	serial.Flush(modlogger)

	// Detect Config Reload
	go func() {
		configReloadedChannel := d.SubscribeToChanges()

		const stopDelay = 50 * time.Millisecond

		for {
			select {
			case <-configReloadedChannel:
				modlogger.Named("Display").Debug("Config Reload Detected")
				serial.Pause()

				cfgDSP.Load()

				loadDSPMapings(modlogger)
				modlogger.Named("Serial").Debug("Flushing")
				serial.Flush(modlogger)
				// let the connection close
				<-time.After(stopDelay)
				//Initalise the Displays

				serial.Start()
			}
		}
	}()

	// go func() {
	// 	sessionReloadedChannel := d.SubscribeToSessionReload()

	// 	for {
	// 		select {
	// 		case <-sessionReloadedChannel:
	// 			serial.Pause()
	// 			modlogger.Named("Display").Debug("Session Reload Detected")
	// 			loadDSPMapings(modlogger)
	// 			serial.Start()
	// 		}
	// 	}
	// }()

	d.Start()

}

func loadDSPMapings(modlogger *zap.SugaredLogger) {
	modlogger = modlogger.Named("Display")

	modlogger.Info("Setting Displays")

	// Create an automap for the sessions
	AutoMap := deejdsp.CreateAutoMap(sliderMap, sessionMap)
	modlogger.Debugf("AutoMaped Sessions: %v", AutoMap)
	//for each screen go and check the config and finaly set the image
	for key, value := range cfgDSP.DisplayMapping {
		serTCA.SelectPort(uint8(key))
		if value != "auto" { // Set to name in the customised image
			fileExsists, _ := serSD.CheckForFile(value)
			if fileExsists {
				serDSP.SetImage(string(value))
				serDSP.DisplayOn()
				modlogger.Debugf("%d: %q", key, value)
			} else {
				modlogger.Debugf("%d: imagefile with name %q does not exsist on remote", key, value)
			}
		} else if len(value) <= 0 { // Turn the display off if nothing is set
			serDSP.DisplayOff()
		} else if value == "auto" { // if its set to auto: Generate a image if it does not exsist and send it to the SD card
			//get the audio session from deej using the AutoMap
			if autoMappedImage, ok := AutoMap[key]; ok {
				SessionAtSlider, _ := sessionMap.Get(autoMappedImage)
				if SessionAtSlider != nil {
					// get the icon path
					iconPathFull := SessionAtSlider[0].GetIconPath()
					icoPathElements := strings.Split(iconPathFull, "\\")
					icoFileName := icoPathElements[len(icoPathElements)-1]
					sdname := deejdsp.CreateFileName(icoFileName)
					// if the program has a iconpath cont
					if len(iconPathFull) > 0 {

						// Check if the file exsits on the card
						pregenerated, _ := serSD.CheckForFile(sdname)

						// Testing to always generate an image
						// pregenerated = false

						// generate a new image if it doesnt exsist
						if !pregenerated {
							slicedIMG, err := deejdsp.GetAndConvertIMG(iconPathFull, 0, cfgDSP.BWThreshold)
							if err != nil {
								modlogger.Errorf("No Image found at the filepath")
								break
							}
							var byteslice []byte
							for _, value := range slicedIMG {
								for _, value2 := range value {
									byteslice = append(byteslice, value2)
								}
							}
							serSD.SendByteSlice(byteslice, sdname)
						}
						serDSP.SetImage(sdname)
						serDSP.DisplayOn()
						modlogger.Debugf("file path of slider %d: %s", key, iconPathFull)
						modlogger.Debugf("%d: program %q localfile %q", key, icoFileName, sdname)
					} else {
						modlogger.Debugf("No Session Mapped for Slider %d", key)
					}
				}
			}
		}
	}
}
