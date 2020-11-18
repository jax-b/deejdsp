package deejdsp

import (
	"errors"
	"strings"
	"time"

	"github.com/jax-b/deej"
	"go.uber.org/zap"
)

// SerialDSP stuct for Serial Dispaly Objects
type SerialDSP struct {
	sio      *deej.SerialIO
	siu      *SerialInUse
	logger   *zap.SugaredLogger
	cmddelay time.Duration
}

// NewSerialDSP Creates a new DSP object
func NewSerialDSP(sio *deej.SerialIO, siu *SerialInUse, logger *zap.SugaredLogger) (*SerialDSP, error) {
	sdlogger := logger.Named("Display")
	serDSP := &SerialDSP{
		sio:    sio,
		siu:    siu,
		logger: sdlogger,
	}
	return serDSP, nil
}

// SetTimeDelay sets the time to delay after a command has been executed
func (serDSP *SerialDSP) SetTimeDelay(delay time.Duration) {
	serDSP.cmddelay = delay
}

// DisplayOn turns the dislpay on
func (serDSP *SerialDSP) DisplayOn() error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}
	if serDSP.siu.ExternalInUse() {
		c := serDSP.siu.JoinLine()
		<-c
		c = nil
	}
	serDSP.siu.PreformingTask()

	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.on")

	if serDSP.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serDSP.cmddelay)
	}

	if resumeAfter {
		serDSP.sio.Start()
	}
	serDSP.siu.Done()
	return nil
}

// DisplayOff turns the dislpay off
func (serDSP *SerialDSP) DisplayOff() error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}
	if serDSP.siu.ExternalInUse() {
		c := serDSP.siu.JoinLine()
		<-c
		c = nil
	}
	serDSP.siu.PreformingTask()

	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.off")

	if serDSP.cmddelay > (time.Microsecond * 1) {
		time.Sleep(serDSP.cmddelay)
	}

	if resumeAfter {
		serDSP.sio.Start()
	}
	serDSP.siu.Done()
	return nil
}

// SetImage Sends the string of the filename for the image selection
func (serDSP *SerialDSP) SetImage(filename string) error {
	resumeAfter := serDSP.sio.IsRunning()

	if serDSP.sio.IsRunning() {
		serDSP.sio.Pause()
	}
	if serDSP.siu.ExternalInUse() {
		c := serDSP.siu.JoinLine()
		<-c
		c = nil
	}
	serDSP.siu.PreformingTask()

	lineChannel := serDSP.sio.ReadLine(serDSP.logger)
	select {
	case <-time.After(350 * time.Millisecond):
		break
	case <-lineChannel:
		break
	}
	serDSP.sio.WriteStringLine(serDSP.logger, "deej.modules.display.setimage")
	time.Sleep(5 * time.Millisecond)
	serDSP.sio.WriteStringLine(serDSP.logger, filename)

Loop:
	for {
		select {
		case <-time.After(350 * time.Millisecond):
			lineChannel = nil

			if resumeAfter {
				serDSP.sio.Start()
			}

			if serDSP.cmddelay > (time.Microsecond * 1) {
				serDSP.logger.Infof("Sleeping for cmd: %s milliseconds", serDSP.cmddelay)
				time.Sleep(serDSP.cmddelay)
			}

			return errors.New("TIMEOUT")
		case msg := <-lineChannel:
			msg = strings.TrimSuffix(msg, "\r\n")
			if msg == "DONE" {
				break Loop
			} else {
				serDSP.logger.Info(msg)
			}
		}
	}

	lineChannel = nil

	if serDSP.cmddelay > (time.Microsecond * 1) {
		// serDSP.logger.Infof("Sleeping for cmd: %s milliseconds", serDSP.cmddelay)
		time.Sleep(serDSP.cmddelay)
	}

	if resumeAfter {
		serDSP.sio.Start()
	}
	serDSP.siu.Done()
	return nil
}
