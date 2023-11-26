package comm

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	fi "gomav/code/trjxtransfer/fileio"
	bf "gomav/code/util/lib"
)

// https://kobatako.hatenablog.com/entry/2017/11/24/210943

type TCPClient struct {
	comm
	destination string
	conn        net.Conn
	connMtx     sync.Mutex
	dbgPrint    *dbgIntervalPrint
}

func CreateTCPClient(destination fi.Address) *TCPClient {

	//buf := make([]byte, 1024)
	log.Printf("Starting TCP Client...(dest=%s:%d)", destination.Address, destination.Port)
	obj := TCPClient{
		destination: fmt.Sprintf("%s:%d", destination.Address, destination.Port),
		dbgPrint:    createDbgIntervalPrint("A->TCPClient", 10),
	}

	Wg.Add(1)
	go obj.runTCPLoop()

	return &obj
}

func (us *TCPClient) runTCPLoop() {
	defer Wg.Done()
	for {
		// TCP接続
		bufferObj := bf.CreateMavBuffer()
		if conn, err := net.Dial("tcp", us.destination); conn != nil && err == nil {
			log.Printf("TCPClient receiveLoop start")
			us.CloseConn()
			us.conn = conn

		receiveLoop:
			for {
				// AutopilotからのTCP受信
				buf := make([]byte, 1024)
				n, err := us.conn.Read(buf)
				if err != nil || n <= 0 {
					us.CloseConn()
					break receiveLoop
				}

				bufferObj.Push(buf[0:n])
				for packet := bufferObj.Pop(); packet != nil; packet = bufferObj.Pop() {
					us.dbgPrint.dbgIntervalPrint()
					mavDebug(packet, 0)
					if us.Pair != nil {
						us.Pair.send(packet)
					}
				}
			}

		} else {
			log.Printf("wait net.Dial e=%v", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func (us *TCPClient) send(data []byte) {
	if us != nil && us.conn != nil {
		if n, err := us.conn.Write(data); err != nil {
			log.Printf("Error. conection close. Write n=%v err=%v", n, err)
			us.CloseConn()
		}
	}
}

func (us *TCPClient) CloseConn() {
	us.connMtx.Lock()
	defer us.connMtx.Unlock()
	if us.conn != nil {
		us.conn.Close()
		us.conn = nil
	}
}
