package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/trajectoryjp/trjx-mavlink-transfer/comm"

	fi "github.com/trajectoryjp/trjx-mavlink-transfer/fileio"
)

func main() {
	fmt.Printf("trjxTransfer v1.2.0\n")
	conf := fi.ReadConfig()

	aircraftID, autopilotSelector, address := parameterSelector(os.Args[1:], conf)

	var trjx, autopilot comm.CommInterface

	switch conf.TRJXSelector {

	case fi.TRJXtAttrGRPC:
		trjx = comm.CreateGRPClient(aircraftID, conf.TRJXgRPC)

	case fi.TRJXAttrUDP:
		trjx = comm.CreateUDPComm(conf.TRJXUDP.MyPort, conf.TRJXUDP.Destination)

	case fi.TRJXAttrTCPClient:

	default:
		log.Fatal("unsupport TRJXSelector")
	}

	//switch conf.AutopilotSelector {
	switch autopilotSelector {
	case fi.AutopilotAttrSerial:
		autopilot = comm.CreateSerial(conf.AutopilotSerial)

	case fi.AutopilotAttrUDP:
		autopilot = comm.CreateUDPComm(conf.AutopilotUDP.MyPort, conf.AutopilotUDP.Destination)

	case fi.AutopilotAttrTCPClient:
		if address != nil {
			autopilot = comm.CreateTCPClient(*address)

		} else {
			autopilot = comm.CreateTCPClient(conf.AutopilotTCPClient)
		}

	case fi.AutopilotAttrTCPServer:
		log.Fatal("unsupport AutopilotAttrTCPServer")

	default:
		log.Fatal("unsupport AutopilotSelector")
	}

	trjx.SetPartner(autopilot)
	autopilot.SetPartner(trjx)

	comm.Wg.Wait()
}

func parameterSelector(params []string, conf *fi.TrjxAircraftConfig) (aircarftID string, autopilotSelector fi.AutopilotAttr, address *fi.Address) {
	st := 0
	aircarftID = conf.AircraftID
	autopilotSelector = conf.AutopilotSelector

	setaircarftID := false

	for _, v := range params {
		switch st {
		case 0:
			switch v {
			case "-h", "-help":
				printHelp()
				os.Exit(0)

			case "-tcp":
				st = 1
				autopilotSelector = fi.AutopilotAttrTCPClient

			default:
				if !setaircarftID {
					aircarftID = v
					setaircarftID = true

				} else {
					fmt.Printf("parameter error:%v\n", params)
					printHelp()
					os.Exit(0)
				}

			}

		case 1:
			if ad := strings.Split(v, ":"); len(ad) == 2 {
				if port, err := strconv.Atoi(ad[1]); err == nil {
					address = &fi.Address{
						Address: ad[0],
						Port:    port,
					}

				} else {
					fmt.Printf("Error port is not number :%v\n", v)
					printHelp()
					os.Exit(0)
				}

			} else {
				fmt.Printf("address[%v] is illegal format\n", v)
				printHelp()
				os.Exit(0)
			}

			st = 0
		}
	}
	return aircarftID, autopilotSelector, address
}

func printHelp() {
	fmt.Print("client <uavid> <address>\n [sample]\n client 87 -tcp 192.168.1.20:5760   <- IPアドレスはAutopilot(SITL等)\n")
}
