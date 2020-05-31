package deejdsp

import (
	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialDSP stuct for Serial Dispaly Objects
type SerialDSP struct {
	sio    *deej.SerialIO
	logger *zap.SugaredLogger
}

// NewSerialDSP Creates a new DSP object
func NewSerialDSP(sio *deej.SerialIO, logger *zap.SugaredLogger) (*SerialDSP, error) {
	sdlogger := logger.Named("display")
	serDSP := &SerialDSP{
		sio:    sio,
		logger: sdlogger,
	}
	return serDSP, nil
}

// DisplayOn turns the dislpay on
func (serDSP *SerialDSP) DisplayOn() error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}

	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.on")

	if resumeAfter {
		serDSP.sio.Start()
	}

	return nil
}

// DisplayOff turns the dislpay off
func (serDSP *SerialDSP) DisplayOff() error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}

	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.off")

	if resumeAfter {
		serDSP.sio.Start()
	}

	return nil
}

// SetImage Sends the string of the filename for the image selection
func (serDSP *SerialDSP) SetImage(filename string) error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}

	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.setimage")
	serDSP.sio.WriteStringLine(serDSP.logger, filename)

	if resumeAfter {
		serDSP.sio.Start()
	}

	return nil
}
