package fileio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config コンフィグ
type TrjxAircraftConfig struct {
	AircraftID string `json:"AircraftID"` // コマンドパラメータ優先
	LogLevel   int    `json:"LogLevel"`

	// TRJX Side
	TRJXSelector TRJXAttr
	// gRPC
	TRJXgRPC Address `json:"TRJXgRPC"`
	// UDP
	TRJXUDP UDPConfig `json:"TRJXUPD"`
	// TCPClient
	TRJXTCPClient Address `json:"TRJXTCPClient"`

	// Autopilot Side
	AutopilotSelector AutopilotAttr
	// Serial
	AutopilotSerial SerialConfig `json:"AutopilotSerial"`
	// UDP
	AutopilotUDP UDPConfig `json:"AutopilotUDP"`
	// TCPClient
	AutopilotTCPClient Address `json:"AutopilotTCPClient"` // 接続先TCPサーバー

}

type TRJXAttr int

const (
	TRJXtAttrGRPC TRJXAttr = iota
	TRJXAttrUDP
	TRJXAttrTCPClient
)

type AutopilotAttr int

const (
	AutopilotAttrSerial AutopilotAttr = iota
	AutopilotAttrUDP
	AutopilotAttrTCPClient
	AutopilotAttrTCPServer
)

type UDPConfig struct {
	MyPort      int     `json:"MyPort"`      // 自身のアドレス
	Destination Address `json:"Destination"` // デフォルト送信先（パケットを受信したら受信元を優先する）
}

type Address struct {
	Address    string `json:"Address"`
	Port       int    `json:"Port"`
	UseTLS     bool   `json:"UseTLS"`
	ServerName string `json:"ServerName"`
}

type SerialConfig struct {
	Name string `json:"Name"`
	Baud int    `json:"Baud"`
	//ParityNone  Parity = 'N'
	//ParityOdd   Parity = 'O'
	//ParityEven  Parity = 'E'
	//ParityMark  Parity = 'M' // parity bit is always 1
	//ParitySpace Parity = 'S' // parity bit is always 0
	Parity string `json:"Parity"`
	// Number of stop bits to use. Default is 1 (1 stop bit).
	//Stop1     StopBits = 1
	//Stop1Half StopBits = 15
	//Stop2     StopBits = 2
	StopBits byte `json:"StopBits"`
	Bit      byte `json:"Bit"`
}

// configデータ
var TrjxAircraftConfigData TrjxAircraftConfig

// ReadConfig setting/config.jsonを読む
func ReadConfig() *TrjxAircraftConfig {
	fmt.Printf("ReadConfig")
	TrjxAircraftConfigData.LogLevel = 3
	TrjxAircraftConfigData.TRJXgRPC.Address = ""
	TrjxAircraftConfigData.TRJXgRPC.Port = 50063
	TrjxAircraftConfigData.TRJXgRPC.UseTLS = false
	TrjxAircraftConfigData.TRJXgRPC.ServerName = ""
	TrjxAircraftConfigData.AutopilotUDP.MyPort = 14551

	contents, err := ioutil.ReadFile("setting/config.json")
	if err == nil {
		error := json.Unmarshal(contents, &TrjxAircraftConfigData)
		if error == nil {
			if TrjxAircraftConfigData.AircraftID == "" {
				fmt.Printf("Config Error AircraftID is NULL")
				os.Exit(0)
				return &TrjxAircraftConfigData
			}

			// 出力選択判定。contentsをmapに展開しなおしてキーの有無を判定
			/*
					TrjxAircraftConfigData.TRJXSelector = TRJXtAttrGRPC
					var trjxkeymap []map[string]interface{}
					json.Unmarshal(contents, &trjxkeymap)
				trjxkeymapLoop:
					for _, v := range trjxkeymap {
						if _, ok := v["TRJXgRPC"]; ok {
							TrjxAircraftConfigData.TRJXSelector = TRJXtAttrGRPC
							break trjxkeymapLoop

						} else if _, ok := v["TRJXUPD"]; ok {
							TrjxAircraftConfigData.TRJXSelector = TRJXAttrUDP
							break trjxkeymapLoop

						} else if _, ok := v["TRJXTCPClient"]; ok {
							TrjxAircraftConfigData.TRJXSelector = TRJXAttrTCPClient
							break trjxkeymapLoop

						}
					}
			*/

			// 入力選択判定。contentsをmapに展開しなおしてキーの有無を判定
			TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrSerial
			TrjxAircraftConfigData.TRJXSelector = TRJXtAttrGRPC
			var keymap map[string]interface{}
			json.Unmarshal(contents, &keymap)
			if _, ok := keymap["TRJXgRPC"]; ok {
				TrjxAircraftConfigData.TRJXSelector = TRJXtAttrGRPC

			} else if _, ok := keymap["TRJXUPD"]; ok {
				TrjxAircraftConfigData.TRJXSelector = TRJXAttrUDP

			} else if _, ok := keymap["TRJXTCPClient"]; ok {
				TrjxAircraftConfigData.TRJXSelector = TRJXAttrTCPClient

			}

			if _, ok := keymap["Serial"]; ok {
				TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrSerial

			} else if _, ok := keymap["AutopilotUDP"]; ok {
				TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrUDP

			} else if _, ok := keymap["AutopilotTCPClient"]; ok {
				TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrTCPClient

			} else if _, ok := keymap["AutopilotTCPServer"]; ok {
				TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrTCPServer
				fmt.Printf("Config not support AutopilotTCPServer")
				os.Exit(0)
			}

			/*
				autopiotkeymapLoop:
						for key, v := range autopiotkeymap {
							if _, ok := v["Serial"]; ok {
								TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrSerial
								break autopiotkeymapLoop

							} else if _, ok := v["AutopilotUDP"]; ok {
								TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrUDP
								break autopiotkeymapLoop

							} else if _, ok := v["AutopilotTCPClient"]; ok {
								TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrTCPClient
								break autopiotkeymapLoop

							} else if _, ok := v["AutopilotTCPServer"]; ok {
								TrjxAircraftConfigData.AutopilotSelector = AutopilotAttrTCPServer
								fmt.Printf("Config not support AutopilotTCPServer")
								os.Exit(0)
								break autopiotkeymapLoop
							}
						}
			*/

			return &TrjxAircraftConfigData
		}
	}
	fmt.Printf("ERROR read setting/config.json:%v", err)
	os.Exit(0)
	return &TrjxAircraftConfigData
}
