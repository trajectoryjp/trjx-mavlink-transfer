package comm

import (
	"log"
	"sync"
	"time"

	"github.com/tarm/serial"

	bf "gomav/code/util/lib"

	fi "github.com/trajectoryjp/trjx-mavlink-transfer/fileio"
)

// https://kobatako.hatenablog.com/entry/2017/11/24/210943
// https://pkg.go.dev/github.com/tarm/serial
// https://github.com/aler9/gomavlib/blob/main/examples/endpoint-serial/main.go
// "/dev/ttyUSB0:57600"

type Serial struct {
	comm
	sconfig  serial.Config
	conn     *serial.Port
	connMtx  sync.Mutex
	dbgPrint *dbgIntervalPrint
}

func CreateSerial(sc fi.SerialConfig) *Serial {

	pt := serial.ParityNone
	switch sc.Parity {
	case "N":
		pt = serial.ParityNone
	case "O":
		pt = serial.ParityOdd
	case "E":
		pt = serial.ParityEven
	case "M":
		pt = serial.ParityMark
	case "S":
		pt = serial.ParitySpace
	}
	sb := serial.Stop1
	switch sc.StopBits {
	case 1, 15, 2:
		sb = serial.StopBits(sc.StopBits)
	}
	obj := Serial{
		sconfig: serial.Config{
			Name:     sc.Name,
			Baud:     sc.Baud,
			Parity:   pt,
			StopBits: sb,
			Size:     sc.Bit,
			//ReadTimeout: time.Millisecond * 100,
		},
		dbgPrint: createDbgIntervalPrint("A->Serial", 10),
	}

	log.Printf("Starting Serial...(%s:%d,parity=%v,stopbits=%v)", obj.sconfig.Name, obj.sconfig.Baud, obj.sconfig.Parity, obj.sconfig.StopBits)
	Wg.Add(1)
	go obj.runSerialLoop()

	return &obj
}

func (cm *Serial) runSerialLoop() {
	defer Wg.Done()
	for {
		// Serial接続
		bufferObj := bf.CreateMavBuffer()
		if conn, err := serial.OpenPort(&cm.sconfig); conn != nil && err == nil {
			log.Printf("Serial receiveLoop start")
			cm.CloseConn()
			cm.conn = conn

		receiveLoop:
			for {
				// AutopilotからのSerial受信
				buf := make([]byte, 1024)
				n, err := cm.conn.Read(buf)
				if err != nil {
					cm.CloseConn()
					break receiveLoop
				}

				if n > 0 {
					bufferObj.Push(buf[0:n])
					for packet := bufferObj.Pop(); packet != nil; packet = bufferObj.Pop() {
						cm.dbgPrint.dbgIntervalPrint()
						mavDebug(packet, 0)
						if cm.Pair != nil {
							cm.Pair.send(packet)
						}
					}
				}
			}

		} else {
			log.Printf("wait serial.OpenPort e=%v", err)
			cm.CloseConn()
		}
		time.Sleep(10 * time.Second)
	}
}

func (cm *Serial) send(data []byte) {
	if cm.conn != nil {
		if n, err := cm.conn.Write(data); err != nil {
			log.Printf("Error. conection close. Write n=%v err=%v", n, err)
			cm.CloseConn()
		}

	} else {
		log.Printf("Serial conn is nil")
	}

}

func (cm *Serial) CloseConn() {
	cm.connMtx.Lock()
	defer cm.connMtx.Unlock()
	if cm.conn != nil {
		cm.conn.Close()
		cm.conn = nil
	}
}
