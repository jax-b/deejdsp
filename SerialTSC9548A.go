package deejdsp

import (
	"errors"
	"strconv"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialTSC strut for serial objects
type SerialTSC struct {
	sio    *deej.SerialIO
	logger *zap.SugaredLogger
}

// NewSerialTSC Creates a new TSC object
func NewSerialTSC(sio *deej.SerialIO, logger *zap.SugaredLogger) (*SerialTSC, error) {
	sdlogger := logger.Named("TSC9548A")
	serTSC := &SerialTSC{
		sio:    sio,
		logger: sdlogger,
	}
	return serTSC, nil
}

// SelectPort Slect the port number on the TSC9548A
// Port can be 0-7
func (serTSC *SerialTSC) SelectPort(PortNumber uint8) error {
	if PortNumber < 0 || PortNumber > 7 {
		return errors.New("Out of bounds")
	}

	resumeAfter := serTSC.sio.IsRunning()

	if serTSC.sio.IsRunning() {
		serTSC.sio.Pause()
	}

	serTSC.sio.WriteStringLine(serTSC.logger, "deej.modules.TSC9548A.select")
	serTSC.sio.WriteStringLine(serTSC.logger, strconv.Itoa(int(PortNumber)))

	if resumeAfter {
		serTSC.sio.Start()
	}

	return nil
}
