package deejdsp

import (
	"bufio"
	"errors"
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
}

// NewSerialSD Creates a new sd object
func NewSerialSD(sio *deej.SerialIO, logger *zap.SugaredLogger) (*SerialSD, error) {
	sdlogger := logger.Named("SD")
	serSD := &SerialSD{
		sio:    sio,
		logger: sdlogger,
	}
	return serSD, nil
}

// SetTimeDelay sets the time to delay after a command has been executed
func (serSD *SerialSD) SetTimeDelay(delay time.Duration) {
	serSD.cmddelay = delay
}

// ListDir lists the dir to logger and returns it as a string
func (serSD *SerialSD) ListDir() (string, error) {
	resumeAfter := serSD.sio.IsRunning()

	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}

	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.list")

	lineChannel := serSD.sio.ReadLine(serSD.logger)

	SerialData := <-lineChannel
	returnText := SerialData

Loop:
	for {
		select {
		case <-time.After(1 * time.Second):
			break Loop
		case SerialData = <-lineChannel:
			if SerialData == "DONE" {
				break Loop
			} else {
				returnText = returnText + SerialData
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

	return returnText, nil
}

// Delete deletes a file off of the SD card
func (serSD *SerialSD) Delete(filename string) error {
	resumeAfter := serSD.sio.IsRunning()

	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}

	filename = strings.ToUpper(filename)

	serSD.logger.Debugf("Deleting %q from the SD Card", filename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.delete")
	serSD.sio.WriteStringLine(serSD.logger, filename)

	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}

	if resumeAfter {
		serSD.sio.Start()
	}

	return nil
}

// SendFile Sends a file to the sd card
func (serSD *SerialSD) SendFile(filepath string, DestFilename string) error {
	resumeAfter := serSD.sio.IsRunning()

	if serSD.sio.IsRunning() {
		serSD.sio.Pause()
	}

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

	n, err := r.Read(b)
	if err == io.EOF {
		return err
	}
	serSD.sio.WriteBytes(serSD.logger, b[0:n])
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

			if err = f.Close(); err != nil {
				return err
			}

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

	if err = f.Close(); err != nil {
		return err
	}

	if serSD.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serSD.cmddelay)
	}

	if resumeAfter {
		serSD.sio.Start()
	}

	return nil
}
