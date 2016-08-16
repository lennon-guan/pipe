package pipe

type WaitIndex struct {
	chs     []chan bool
	allDone chan bool
}

func NewWaitIndex(total int) *WaitIndex {
	chs := make([]chan bool, total)
	for i := 0; i < total; i++ {
		chs[i] = make(chan bool, 1)
	}
	return &WaitIndex{
		chs:     chs,
		allDone: make(chan bool, 1),
	}
}

func (wait *WaitIndex) Wait(index int) {
	if index >= 0 && index < len(wait.chs) {
		<-wait.chs[index]
	}
}

func (wait *WaitIndex) Done(index int) {
	if index >= 0 && index < len(wait.chs) {
		wait.chs[index] <- true
		if index == len(wait.chs)-1 {
			wait.allDone <- true
		}
	}
}

func (wait *WaitIndex) WaitAndClose() {
	<-wait.allDone
	for _, ch := range wait.chs {
		close(ch)
	}
}
