package deejdsp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialSD strut for serial objects
type SerialSD struct {
	sio      *deej.SerialIO
	logger   *zap.SugaredLogger
	cmddelay time.Duration
	verbose  bool
	siu      *SerialInUse
}

// NewSerialSD Creates a new sd object
func NewSerialSD(sio *deej.SerialIO, siu *SerialInUse, logger *zap.SugaredLogger, verbose bool) (*SerialSD, error) {
	sdlogger := logger.Named("SD")
	serSD := &SerialSD{
		sio:     sio,
		logger:  sdlogger,
		verbose: verbose,
		siu:     siu,
	}
	return serSD, nil
}

// CheckForFile Checks if a file exsists on the SD card
func (serSD *SerialSD) CheckForFile(filename string) (bool, error) {
	filelist, err := serSD.ListDir()
	if err != nil {
		return false, err
	}
	return serSD.CheckForFileLOAD(filename, filelist)
}

// CheckForFile Checks if a file exsists on the SD card
func (serSD *SerialSD) CheckForFileLOAD(filename string, filelist []string) (bool, error) {
	filename = strings.ToLower(filename)
	for _, value := range filelist {
		value = strings.ToLower(value)
		if serSD.verbose {
			serSD.logger.Debugf("%s compared to %s", value, filename)
		}
		if strings.EqualFold(value, filename) {
			return true, nil
		}
	}
	return false, nil
}

// SetTimeDelay sets the time to delay after a command has been executed
func (serSD *SerialSD) SetTimeDelay(delay time.Duration) {
	serSD.cmddelay = delay
}

// ListDir lists the dir to logger and returns it as a string
func (serSD *SerialSD) ListDir() ([]string, error) {
	resumeAfter := serSD.sio.IsRunning()
	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}
	if serSD.siu.ExternalInUse() {
		c := serSD.siu.JoinLine()
		fmt.Print("Here")
		<-c
		c = nil
	}
	serSD.siu.PreformingTask()

	var returnText []string
	var SerialData string
	lineChannel := serSD.sio.ReadLine(serSD.logger)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.list")
	time.Sleep(time.Millisecond * 50)
Loop:
	for {
		select {
		case <-time.After(time.Second * 1):
			break Loop
		case SerialData = <-lineChannel:
			SerialData = strings.Replace(SerialData, "\n", "", -1)
			SerialData = strings.Replace(SerialData, "\r", "", -1)
			if SerialData == "DONE" {
				break Loop
			} else {
				returnText = append(returnText, SerialData)
			}
		}
	}

	lineChannel = nil
	fmt.Println(returnText)
	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}

	if resumeAfter {
		serSD.sio.Start()
	}
	serSD.siu.Done()
	return returnText, nil
}

// Delete deletes a file off of the SD card
func (serSD *SerialSD) Delete(filename string) error {
	resumeAfter := serSD.sio.IsRunning()
	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}
	if serSD.siu.ExternalInUse() {
		c := serSD.siu.JoinLine()
		<-c
		c = nil
	}
	serSD.siu.PreformingTask()

	filename = strings.ToUpper(filename)

	serSD.logger.Debugf("Deleting %q from the SD Card", filename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.delete")
	serSD.sio.WriteStringLine(serSD.logger, filename)

	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}
	serSD.siu.Done()
	if resumeAfter {
		serSD.sio.Start()
	}
	serSD.siu.Done()
	return nil
}

// SendFile Sends a file to the sd card
func (serSD *SerialSD) SendFile(filepath string, DestFilename string) error {
	resumeAfter := serSD.sio.IsRunning()

	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}
	if serSD.siu.ExternalInUse() {
		c := serSD.siu.JoinLine()
		<-c
		c = nil
	}
	serSD.siu.PreformingTask()

	serSD.logger.Debugf("Sending %q to the SD Card with %q as the file name", filepath, DestFilename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.send")

	serSD.sio.WriteStringLine(serSD.logger, DestFilename)

	finfo, err := os.Stat(filepath)
	fsize := finfo.Size()

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(f)
	b := make([]byte, fsize)

	_, err = r.Read(b)
	if err == io.EOF {
		return err
	}
	//send each byte with a small delay between each byte
	for _, value := range b {
		var valsl []byte
		valsl = append(valsl, value)
		serSD.sio.WriteBytes(serSD.logger, valsl)
		time.Sleep(time.Millisecond * 1)
	}
	serSD.sio.WriteStringLine(serSD.logger, "EOF")
	// create line channel
	lineChannel := serSD.sio.ReadLine(serSD.logger)

	// Watch for done message since this is time intensive
	// If it takes to long exit
Loop:
	for {
		select {
		case <-time.After(750 * time.Millisecond):
			lineChannel = nil

			if err = f.Close(); err != nil {
				return err
			}

			if resumeAfter {
				serSD.sio.Start()
			}
			return errors.New("Timeout (waiting for arduino)")
		case msg := <-lineChannel:
			msg = strings.TrimSuffix(msg, "\r\n")
			msg = strings.TrimPrefix(msg, " ")
			if msg == "DONE" {
				break Loop
			} else if msg == "" {
			} else {
				serSD.logger.Info(msg)
			}
		}
	}

	lineChannel = nil

	if err = f.Close(); err != nil {
		return err
	}

	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}

	if resumeAfter {
		serSD.sio.Start()
	}
	serSD.siu.Done()
	return nil
}

// SendByteSlice Sends a file to the sd card
func (serSD *SerialSD) SendByteSlice(byteslice []byte, DestFilename string) error {
	resumeAfter := serSD.sio.IsRunning()

	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}
	if serSD.siu.ExternalInUse() {
		c := serSD.siu.JoinLine()
		<-c
		c = nil
	}
	serSD.siu.PreformingTask()

	serSD.logger.Debugf("Sending bytes to the SD Card with %q as the file name", DestFilename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.send")

	serSD.sio.WriteStringLine(serSD.logger, DestFilename)

	serSD.sio.WriteBytes(serSD.logger, byteslice)
	serSD.sio.WriteStringLine(serSD.logger, "EOF")

	//clear status messages
	lineChannel := serSD.sio.ReadLine(serSD.logger)

	// Watch for done message since this is time intensive
	// If it takes to long exit
Loop:
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			lineChannel = nil

			if resumeAfter {
				serSD.sio.Start()
			}
			return errors.New("Timeout (waiting for arduino)")
		case msg := <-lineChannel:
			msg = strings.TrimSuffix(msg, "\r\n")
			if msg == "DONE" {
				break Loop
			} else if msg == "" {
			} else {
				serSD.logger.Info(msg)
			}
		}
	}

	lineChannel = nil

	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}

	if resumeAfter {
		serSD.sio.Start()
	}
	serSD.siu.Done()
	return nil
}
