package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/jax-b/deej"
	"github.com/jax-b/deejdsp"
	"github.com/jax-b/iconfinderapi"
	"github.com/sqweek/dialog"
	"go.uber.org/zap"
)

var (
	gitCommit  string
	versionTag string
	buildType  string

	verbose       bool
	useIconFinder bool

	d          *deej.Deej
	cfgDSP     *deejdsp.DSPCanonicalConfig
	serSD      *deejdsp.SerialSD
	serTCA     *deejdsp.SerialTCA
	serDSP     *deejdsp.SerialDSP
	siumonitor *deejdsp.SerialInUse
	sessionMap *deej.SessionMap
	sliderMap  *deej.SliderMap
	icofdrapi  *iconfinderapi.Iconfinder

	crntDSPimg map[int]string
)

const stopDelay = 50 * time.Millisecond

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

	// Create IconFinderAPI
	if strings.EqualFold(cfgDSP.IconFinderDotComAPIKey, "silent") {
		modlogger.Info("iconfinder.com apikey not set: in order to use online icons please enter a icon finder api key")
		useIconFinder = false
	} else if len(cfgDSP.IconFinderDotComAPIKey) > 0 && !strings.EqualFold(cfgDSP.IconFinderDotComAPIKey, "example") {
		icofdrapi = iconfinderapi.NewIconFinder(cfgDSP.IconFinderDotComAPIKey)
		useIconFinder = true
	} else {
		useIconFinder = false
		modlogger.Info("iconfinder.com apikey not set: in order to use online icons please enter a icon finder api key")
		d.Notifier.Notify("iconfinder.com apikey not set", "in order to use online icons please enter a icon finder api key")
	}

	if cfgDSP.StartupDelay > 0 {
		modlogger.Debugf("Sleeping for controller startup: %s milliseconds", cfgDSP.StartupDelay)
		time.Sleep(time.Duration(cfgDSP.StartupDelay) * time.Millisecond)
	}

	//Set up all modules

	siumonitor = deejdsp.NewSerialInUse()

	serSD, err = deejdsp.NewSerialSD(serial, siumonitor, modlogger, verbose)
	serTCA, err = deejdsp.NewSerialTCA(serial, siumonitor, modlogger)
	serDSP, err = deejdsp.NewSerialDSP(serial, siumonitor, modlogger)

	crntDSPimg = make(map[int]string)

	// Tray Menu item: Send Image
	go func() {
		menuItemChan := d.AddMenuItem("Send Image", "Send A image file to the internal SD card")
		menuItem := <-menuItemChan
		for {
			<-menuItem.ClickedCh
			filename, err := dialog.File().Filter("ByteImage", "b").Title("Send Image").Load()
			if err != nil {
				break
			}
			PathElements := strings.Split(filename, "\\")
			sdFilename := PathElements[len(PathElements)-1]
			serSD.SendFile(filename, sdFilename)
			dialog.Message("%s", "File Transfer done").Title("Send Image").Info()
		}
	}()
	time.Sleep(2 * time.Millisecond)
	// Tray Menu item: List Files
	go func() {
		menuItemChan := d.AddMenuItem("List Files", "List the files on the sd card")
		menuItem := <-menuItemChan
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
	time.Sleep(2 * time.Millisecond)
	// Tray Menu Item : Turn off displays
	go func() {
		menuItemChan := d.AddMenuItem("Displays Off", "If the config reloads it will turn off all the displays")
		menuItem := <-menuItemChan
		for {
			<-menuItem.ClickedCh
			resumeAfter := serial.IsRunning()
			modlogger.Named("Displays").Debug("Turning Off Displays")
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
	time.Sleep(2 * time.Millisecond)
	// Tray Menu Item : Reconfig Displays
	go func() {
		menuItemChan := d.AddMenuItem("Displays Reload", "Turns on the displays to the images in the config")
		menuItem := <-menuItemChan
		for {
			<-menuItem.ClickedCh
			resumeAfter := serial.IsRunning()

			if serial.IsRunning() {
				serial.Pause()
			}

			loadDSPMapings(modlogger)

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

	// Detect Config Reload
	go func() {
		configReloadedChannel := d.SubscribeToChanges()

		for {
			select {
			case <-configReloadedChannel:
				modlogger.Named("Display").Debug("Config Reload Detected")
				serial.Pause()

				cfgDSP.Load()

				// Update IconFinderAPIkey
				if strings.EqualFold(cfgDSP.IconFinderDotComAPIKey, "silent") {
					modlogger.Info("iconfinder.com apikey not set: in order to use online icons please enter a icon finder api key")
					useIconFinder = false
				} else if len(cfgDSP.IconFinderDotComAPIKey) > 0 && !strings.EqualFold(cfgDSP.IconFinderDotComAPIKey, "example") {
					icofdrapi.ChangeAPIKey(cfgDSP.IconFinderDotComAPIKey)
					useIconFinder = true
				} else {
					useIconFinder = false
					modlogger.Info("iconfinder.com apikey not set: in order to use online icons please enter a icon finder api key")
				}

				sessionMap = d.GetSessionMap()
				sliderMap = d.GetSliderMap()
				loadDSPMapings(modlogger)

				modlogger.Named("Serial").Debug("Flushing")
				serial.Flush(modlogger)
				// let the connection close
				<-time.After(stopDelay)

				serial.Start()
			}
		}
	}()

	go func() {
		sessionReloadedChannel := d.SubscribeToSessionReload()
		configReloadedChannel := d.SubscribeToChanges()
		// Wait till after startup
		time.Sleep(15 * time.Second)
		// Clear any reloads that were triggerd
	Loop1:
		for {
			select {
			case <-sessionReloadedChannel:
			case <-time.After(5 * time.Millisecond):
				break Loop1
			}
		}

		for {
			switch {
			case <-sessionReloadedChannel:
				serial.Pause()
				modlogger.Named("Display").Debug("Session Reload Detected")
				loadDSPMapings(modlogger)
				serial.Start()

				// Minimum deley bettween session reloads for serial
				time.Sleep(1 * time.Second)
				// Clear any reloads that were triggerd
			Loop2:
				for {
					select {
					case <-sessionReloadedChannel:
					case <-time.After(20 * time.Millisecond):
						break Loop2
					}
				}
			case <-configReloadedChannel: //If a session reloads was triggerd by a config reload ignore it and clear the states: Fixes a bug where there was a constant display refresh after config reload
				time.Sleep(15 * time.Millisecond)
			Loop3:
				for {
					select {
					case <-sessionReloadedChannel:
					case <-time.After(20 * time.Millisecond):
						break Loop3
					}
				}
			}
		}
	}()

	serial.Flush(modlogger)
	// let the connection close
	<-time.After(stopDelay)
	d.Start()

}

func loadDSPMapings(modlogger *zap.SugaredLogger) {
	modlogger = modlogger.Named("Display")

	modlogger.Info("Setting Displays")

	// Create an automap for the sessions
	AutoMap := deejdsp.CreateAutoMap(sliderMap, sessionMap)
	modlogger.Debugf("AutoMaped Sessions: %v", AutoMap)
	sdfiles, _ := serSD.ListDir()
	//for each screen go and check the config and finaly set the image
	for key, value := range cfgDSP.DisplayMapping {
		serTCA.SelectPort(uint8(key))
		if value != "auto" { // Set to name in the customised image
			if value != crntDSPimg[key] {
				fileExsists, _ := serSD.CheckForFileLOAD(value, sdfiles)
				if fileExsists {
					serDSP.SetImage(string(value))
					modlogger.Debugf("%d: %q", key, value)
					crntDSPimg[key] = value
				} else {
					modlogger.Debugf("%d: imagefile with name %q does not exsist on remote", key, value)
				}
			}
			serDSP.DisplayOn()
		} else if len(value) <= 0 { // Turn the display off if nothing is set
			serDSP.DisplayOff()
		} else if value == "auto" { // if its set to auto: Generate a image if it does not exsist and send it to the SD card
			//get the audio session from deej using the AutoMap
			if autoMappedImage, ok := AutoMap[key]; ok {
				programname := strings.Split(autoMappedImage, ".")[0]
				sdname := deejdsp.CreateFileName(programname)

				// Check if the file exsits on the card
				pregenerated, _ := serSD.CheckForFileLOAD(sdname, sdfiles)
				customImage, _ := serSD.CheckForFileLOAD(programname+".b", sdfiles)
				if customImage == false {
					customImage, _ = serSD.CheckForFileLOAD(strings.ToLower(programname)+".b", sdfiles)
				}
				// generate a new image if it doesnt exsist
				if !pregenerated && useIconFinder && !customImage {
					//Get Icon from API and convert it to byteslices
					qualifiedico, err := deejdsp.GetIconFromAPI(icofdrapi, programname)
					if err != nil {
						modlogger.Named("Display").Errorf("Could not get image from API, try generating your own image insted for %s: Error Text %s", programname, err.Error())
					} else {
						slicedIMG, err := deejdsp.ConvertImage(qualifiedico, 0, cfgDSP.BWThreshold)
						if err != nil {
							modlogger.Errorf("No Image found in qualifiedico")
							break
						}
						// Convert to a single long slice
						var byteslice []byte
						for _, value := range slicedIMG {
							for _, value2 := range value {
								byteslice = append(byteslice, value2)
							}
						}
						// Send Slice to the SD card
						serSD.SendByteSlice(byteslice, sdname)
						sdfiles = append(sdfiles, sdname)
						// Store the current mapping
						crntDSPimg[key] = sdname
						serDSP.SetImage(sdname)
						modlogger.Debugf("%d: program %q localfile %q", key, programname, sdname)
					}
				} else {
					if customImage {
						if crntDSPimg[key] != programname+".b" {
							crntDSPimg[key] = programname + ".b"
							serDSP.SetImage(programname + ".b")
							modlogger.Debugf("%d: program %q localfile %q", key, programname, programname+".b")
						}
					} else {
						if crntDSPimg[key] != sdname {
							crntDSPimg[key] = sdname
							serDSP.SetImage(sdname)
							modlogger.Debugf("%d: program %q localfile %q", key, programname, sdname)
						}
					}
				}
				serDSP.DisplayOn()
			} else {
				serDSP.DisplayOff()
			}
		}
	}
}
