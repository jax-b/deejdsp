package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/jax-b/deej"
	"github.com/jax-b/deejdsp"
	"go.uber.org/zap"
)

var (
	gitCommit  string
	versionTag string
	buildType  string
	d          *deej.Deej
	cfgDSP     *deejdsp.DSPCanonicalConfig
	serSD      *deejdsp.SerialSD
	serTCA     *deejdsp.SerialTCA
	serDSP     *deejdsp.SerialDSP
	sessionMap *deej.SessionMap
	sliderMap  *deej.SliderMap
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

	if err = d.Initialize(); err != nil {
		named.Fatalw("Failed to initialize deej", "error", err)
	}

	modlogger := d.NewNammedLogger("module")
	serial := d.GetSerial()
	// Load Config
	cfgDSP, err = deejdsp.NewDSPConfig(modlogger)
	cfgDSP.Load()

	if cfgDSP.StartupDelay > 0 {
		modlogger.Infof("Sleeping for controller startup: %s milliseconds", cfgDSP.StartupDelay)
		time.Sleep(time.Duration(cfgDSP.StartupDelay) * time.Millisecond)
	}

	//Set up all modules
	serSD, err = deejdsp.NewSerialSD(serial, modlogger)
	serTCA, err = deejdsp.NewSerialTCA(serial, modlogger)
	serDSP, err = deejdsp.NewSerialDSP(serial, modlogger)

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
	modlogger.Infof("AutoMaped Sessions: %v", AutoMap)
	//for each screen go and check the config and finaly set the image
	for key, value := range cfgDSP.DisplayMapping {
		serTCA.SelectPort(uint8(key))
		if value != "auto" { // Set to name in the customised image
			fileExsists, _ := serSD.CheckForFile(value)
			if fileExsists {
				serDSP.SetImage(string(value[0]))
				serDSP.DisplayOn()
				modlogger.Infof("%d: %q", key, value)
			} else {
				modlogger.Infof("%d: imagefile with name %q does not exsist on remote", key, value)
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
					modlogger.Infof("file path of slider %d: %s", key, iconPathFull)
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
							slicedIMG := deejdsp.GetAndConvertIMG(iconPathFull, 0, cfgDSP.BWThreshold)
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
						modlogger.Infof("%d: program %q localfile %q", key, icoFileName, sdname)
					} else {
						modlogger.Infof("No Session Mapped for Slider %d", key)
					}
				}
			}
		}
	}
}
