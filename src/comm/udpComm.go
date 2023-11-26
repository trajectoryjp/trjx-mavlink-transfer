package comm

import (
	"log"
	"net"

	fi "gomav/code/trjxtransfer/fileio"
	bf "gomav/code/util/lib"
)

// https://kobatako.hatenablog.com/entry/2017/11/24/210943

type UDPComm struct {
	comm
	distination *net.UDPAddr
	conn        *net.UDPConn
	dbgPrint    *dbgIntervalPrint
	//gRPCClient  *GRPClient
}

func CreateUDPComm(myPort int, destination fi.Address) *UDPComm {

	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: myPort,
	}
	updLn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalln(err)
	}

	//buf := make([]byte, 1024)
	log.Printf("Starting UDP Server...(myPort=%v dest=%s:%d)", myPort, destination.Address, destination.Port)

	distination := &net.UDPAddr{
		//IP:   net.ParseIP("192.168.1.24"),
		//Port: 14550,
		IP:   net.ParseIP(destination.Address),
		Port: destination.Port,
	}

	obj := UDPComm{
		distination: distination,
		conn:        updLn,
		dbgPrint:    createDbgIntervalPrint("Autopilot->TRJX", 10),
	}

	Wg.Add(1)
	go obj.runUDPLoop()

	return &obj
}

/*
func (us *UDPComm) SetPartner(p CommInterface) {
	us.Pair = p
}
*/

func (us *UDPComm) runUDPLoop() {
	defer Wg.Done()
	log.Print("UDP receiveLoop start")

	bufferObj := bf.CreateMavBuffer()
	for {
		// AutopilotからのUDP受信
		buf := make([]byte, 1024)
		n, s, err := us.conn.ReadFromUDP(buf)
		if err != nil || s == nil || n <= 0 {
			log.Fatalln(err)
		}

		bufferObj.Push(buf[0:n])
		for packet := bufferObj.Pop(); packet != nil; packet = bufferObj.Pop() {
			us.dbgPrint.dbgIntervalPrint()
			mavDebug(packet, 0)

			us.distination = s
			if us.Pair != nil {
				us.Pair.send(packet)
			}
		}

	}
}

func (us *UDPComm) send(data []byte) {
	if n, err := us.conn.WriteToUDP(data, us.distination); err != nil {
		log.Printf("WriteToUDP n=%v err=%v", n, err)
	}
}
