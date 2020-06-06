package deejdsp

import (
	"errors"
	"strconv"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialTCA strut for serial objects
type SerialTCA struct {
	sio    *deej.SerialIO
	logger *zap.SugaredLogger
}

// NewSerialTCA Creates a new TCA object
func NewSerialTCA(sio *deej.SerialIO, logger *zap.SugaredLogger) (*SerialTCA, error) {
	sdlogger := logger.Named("TCA9548A")
	serTCA := &SerialTCA{
		sio:    sio,
		logger: sdlogger,
	}
	return serTCA, nil
}

// SelectPort Slect the port number on the TCA9548A
// Port can be 0-7
func (serTCA *SerialTCA) SelectPort(PortNumber uint8) error {
	if PortNumber < 0 || PortNumber > 7 {
		return errors.New("Out of bounds")
	}

	resumeAfter := serTCA.sio.IsRunning()

	if serTCA.sio.IsRunning() {
		serTCA.sio.Pause()
	}

	serTCA.sio.WriteStringLine(serTCA.logger, "deej.modules.TCA9548A.select")
	serTCA.sio.WriteStringLine(serTCA.logger, strconv.Itoa(int(PortNumber)))

	if resumeAfter {
		serTCA.sio.Start()
	}

	return nil
}
