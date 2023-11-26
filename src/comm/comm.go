package comm

import "sync"

type CommInterface interface {
	send(buffer []byte)
	SetPartner(p CommInterface)
}

type comm struct {
	Pair     CommInterface
	dbgPrint *dbgIntervalPrint
}

var Wg sync.WaitGroup

func (us *comm) SetPartner(p CommInterface) {
	us.Pair = p
}
