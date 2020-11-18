package deejdsp

type SerialInUse struct {
	funcsWaiting            []chan bool
	ExternalFuncUsingSerial bool
}

func NewSerialInUse() *SerialInUse {
	siu := &SerialInUse{
		funcsWaiting: make([]chan bool, 1),
	}
	return siu
}

func (siu *SerialInUse) JoinLine() chan bool {
	c := make(chan bool)
	siu.funcsWaiting = append(siu.funcsWaiting, c)

	return c
}

func (siu *SerialInUse) ExternalInUse() bool {
	if siu.funcsWaiting[0] == nil {
		return false
	}
	return true
}
func (siu *SerialInUse) PreformingTask() {
	siu.ExternalFuncUsingSerial = true
}
func (siu *SerialInUse) Done() {
	if siu.funcsWaiting[0] != nil {
		// System Uses a FIFO system for controlling external modules access to serial
		var functonotifty chan bool
		if len(siu.funcsWaiting) > 1 {
			functonotifty = siu.funcsWaiting[0]
			copy(siu.funcsWaiting[0:], siu.funcsWaiting[1:]) // Shift a[i+1:] left one index.
			siu.funcsWaiting[len(siu.funcsWaiting)-1] = nil  // Erase last element
			siu.funcsWaiting = siu.funcsWaiting[:len(siu.funcsWaiting)-1]
		} else if len(siu.funcsWaiting) == 1 {
			functonotifty = siu.funcsWaiting[0]
			siu.funcsWaiting[0] = nil
		}
		functonotifty <- true
	} else {
		siu.ExternalFuncUsingSerial = false
	}
}
