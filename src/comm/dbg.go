package comm

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v2/pkg/dialect"
	"github.com/bluenviron/gomavlib/v2/pkg/dialects/all"

	//"github.com/bluenviron/gomavlib/v2/pkg/parser"
	"github.com/bluenviron/gomavlib/v2/pkg/frame"
)

var msgMap = map[uint32]bool{51: true, 47: true, 44: true, 40: true, 77: true, 39: true, 73: true}

//LOCAL_POSITION_NED ( #32 )
//MISSION_REQUEST_INT ( #51 )
//MISSION_ITEM ( #39 )
//MISSION_ITEM_INT ( #73 )

//var msgMap = map[uint32]bool{}

// var dialectDE *dialect.DecEncoder
var dialectRW *dialect.ReadWriter
var once sync.Once

func mavDebugInit() {
	var err error
	//dialectDE, err = dialect.NewDecEncoder(all.Dialect) // この処理は非常に重い
	dialectRW, err = dialect.NewReadWriter(all.Dialect)
	if err != nil {
		log.Fatal("mavDebug NewDecEncoder e=%v", err)
	}
}
func mavDebug(msg []byte, dir int) {

	once.Do(mavDebugInit)

	inBuf := bytes.NewBuffer(msg)
	//reader, err := parser.NewReader(parser.ReaderConf{
	reader, err := frame.NewReader(frame.ReaderConf{
		Reader: inBuf,
		//DialectDE: dialectDE,
		DialectRW: dialectRW,
	})
	if err != nil {
		panic(err)
	}

	dirTxt := "A>T" // 0
	if dir != 0 {
		dirTxt = "A<T" // 1
	}
	// read a message, encapsulated in a frame
	frame, err := reader.Read()
	if err != nil {
		fmt.Printf("mavDebug read e=%v", err)
		//panic(err)
	}
	if frame != nil {
		mav := frame.GetMessage()
		mid := mav.GetID()
		_, ok := msgMap[mid]
		if ok || len(msgMap) == 0 {
			switch v := mav.(type) {
			case *all.MessageMissionCount:
				log.Printf("%v:MissionCount:sys=%v cmp=%v cnt=%v (typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Count, v.MissionType)

			case *all.MessageMissionRequestInt:
				log.Printf("%v:MissionReqInt:sys=%v cmp=%v seq=%v (typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Seq, v.MissionType)

			case *all.MessageMissionRequest:
				log.Printf("%v:MissionReq:sys=%v cmp=%v seq=%v (typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Seq, v.MissionType)

			case *all.MessageMissionItem:
				log.Printf("%v:MissionItem:sys=%v cmp=%v seq=%v cmd=%d [%v,%v,%v](typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Seq, v.Command, v.X, v.Y, v.Z, v.MissionType)

			case *all.MessageMissionItemInt:
				log.Printf("%v:MissionItemInt:sys=%v cmp=%v seq=%v cmd=%d [%v,%v,%v](typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Seq, v.Command, v.X, v.Y, v.Z, v.MissionType)

			case *all.MessageMissionAck:
				log.Printf("%v:MissionAck:sys=%v cmp=%v result=%d (typ=%v)\n", dirTxt, v.TargetSystem, v.TargetComponent, v.Type, v.MissionType)

			default:
				log.Printf("%s:[%v] %+v\n", dirTxt, mid, frame)
			}
			for _, v := range msg {
				fmt.Printf("%02x ", v)
			}
			fmt.Printf("\n")

		}

	} else {
		fmt.Printf("%s: frame is nil\n", dirTxt)
	}

}

type dbgIntervalPrint struct {
	//watchTime int64
	//start      bool
	//startMsg   string
	//timeoutMsg string
	msg string
	//event   chan struct{}
	counter int
}

func createDbgIntervalPrint(msg string, timerSec int) *dbgIntervalPrint {
	obj := dbgIntervalPrint{
		//watchTime: timerSec,
		//event: make(chan struct{}),
		msg: msg,
		//startMsg:   startMsg,
		//timeoutMsg: timeoutMsg,
	}

	go obj.startLoop(timerSec)
	return &obj
}

func (dp *dbgIntervalPrint) startLoop(timerSec int) {
	timer := time.NewTimer(time.Duration(timerSec) * time.Second)
	for {
		<-timer.C
		log.Printf("%s:%d", dp.msg, dp.counter)
		dp.counter = 0
		timer.Reset(time.Duration(timerSec) * time.Second)
		/*
			select {
			case <-timer.C:
				log.Print("%s:%d", dp.msg, dp.counter)
				dp.counter = 0
				timer.Reset(time.Duration(timerSec) * time.Second)
			}
		*/
	}
	/*
		for {
			select {
			case <-timer.C:
				fmt.Print(dp.timeoutMsg + "\n")
				dp.start = false

			case <-dp.event:
				if !dp.start {
					fmt.Print(dp.startMsg + "\n")
					dp.start = true
				}
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(time.Duration(timerSec) * time.Second)
			}
		}
	*/
}

func (dp *dbgIntervalPrint) dbgIntervalPrint() {
	dp.counter++
}
