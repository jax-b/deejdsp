package deejdsp

import (
	"errors"
	"os"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialSD strut for serial objects
type SerialSD struct {
	sio    *deej.SerialIO
	logger *zap.SugaredLogger
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

// ListDir lists the dir to logger and returns it as a string
func (serSD *SerialSD) ListDir() (string, error) {
	serSD.sio.Pause()
	serSD.logger.Info("SDCard File list")
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.list")
	SerialData := <-serSD.sio.ReadLine(serSD.logger)
	returnText := SerialData
	for SerialData != "" {
		serSD.logger.Info(SerialData)
		SerialData = <-serSD.sio.ReadLine(serSD.logger)
		returnText = returnText + SerialData
	}
	serSD.sio.Start()
	return returnText, nil
}

// Delete deletes a file off of the SD card
func (serSD *SerialSD) Delete(filename string) (string, error) {
	serSD.sio.Pause()

	serSD.logger.Infof("Deleting %q from the SD Card", filename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.delete")
	serSD.sio.WriteStringLine(serSD.logger, filename)

	success, cmdKey := serSD.sio.WaitFor(serSD.logger, "FILEDELETED")

	if success == false {
		serSD.logger.Errorw("Failed to delete file", filename)
		return cmdKey, errors.New("Cannot delete file")
	}

	serSD.sio.Start()
	return cmdKey, nil
}

// SendFile Sends a file to the sd card
func (serSD *SerialSD) SendFile(filepath string, DestFilename string) error {
	serSD.sio.Pause()

	serSD.logger.Infof("Sending %q to the SD Card with %q as the file name", filepath, DestFilename)
	serSD.sio.WriteStringLine(serSD.logger, "deej.modules.sd.send")
	serSD.sio.WriteStringLine(serSD.logger, DestFilename)

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	b1 := make([]byte, 1)
	n1, err := f.Read(b1)
	for n1 > 0 {
		serSD.sio.WriteBytes(serSD.logger, b1)
		n1, err = f.Read(b1)
	}
	serSD.sio.WriteStringLine(serSD.logger, "EOF")

	serSD.sio.Start()

	return nil
}
