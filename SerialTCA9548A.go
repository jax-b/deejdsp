package deejdsp

import (
	"errors"
	"strconv"
	"time"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialTCA strut for serial objects
type SerialTCA struct {
	sio      *deej.SerialIO
	siu      *SerialInUse
	logger   *zap.SugaredLogger
	cmddelay time.Duration
}

// NewSerialTCA Creates a new TCA object
func NewSerialTCA(sio *deej.SerialIO, siu *SerialInUse, logger *zap.SugaredLogger) (*SerialTCA, error) {
	sdlogger := logger.Named("TCA9548A")
	serTCA := &SerialTCA{
		sio:    sio,
		siu:    siu,
		logger: sdlogger,
	}
	return serTCA, nil
}

// SetTimeDelay sets the time to delay after a command has been executed
func (serTCA *SerialTCA) SetTimeDelay(delay time.Duration) {
	serTCA.cmddelay = delay
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

	if serTCA.siu.ExternalInUse() {
		c := serTCA.siu.JoinLine()
		<-c
		c = nil
	}
	serTCA.siu.PreformingTask()

	serTCA.sio.WriteStringLine(serTCA.logger, "deej.modules.TCA9548A.select")
	serTCA.sio.WriteStringLine(serTCA.logger, strconv.Itoa(int(PortNumber)))

	if serTCA.cmddelay > (time.Microsecond * 1) {
		// serTCA.logger.Infof("Sleeping for cmd: %s milliseconds", serTCA.cmddelay)
		time.Sleep(serTCA.cmddelay)
	}

	if resumeAfter {
		serTCA.sio.Start()
	}
	serTCA.siu.Done()
	return nil
}
