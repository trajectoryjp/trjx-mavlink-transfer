package comm

import (
	"io"
	"log"
	"time"

	pb "github.com/trajectoryjp/trjx-vehicle-api/proto_go/trjxmavlink"

	fi "github.com/trajectoryjp/trjx-mavlink-transfer/fileio"
	tj "github.com/trajectoryjp/trjx-mavlink-transfer/trjxcomm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPClient struct {
	comm
	stream pb.TrjxMavlinkService_CommunicateOnMavlinkClient
	//dbgPrint  *dbgIntervalPrint
	//UDPServer *UDPServer
}

func CreateGRPClient(aircraft string, address fi.Address) *GRPClient {
	log.Println("Starting gRPC Client...")

	obj := GRPClient{
		comm: comm{
			dbgPrint: createDbgIntervalPrint("TRJX->Autopilot", 10),
		},
	}

	Wg.Add(1)
	go obj.runGRPCLoop(aircraft, address)

	return &obj
}

func (us *GRPClient) runGRPCLoop(aircraft string, address fi.Address) {
	defer Wg.Done()
	for {
		// stream接続
		if us.stream = tj.OpenCommunication(aircraft, address); us.stream != nil { // Login/CommunicateOnMavlink
			log.Printf("gRPC receiveLoop start uavid=%v_", aircraft)
		steamLoop:
			for {
				if rcv, err := us.stream.Recv(); err == nil {
					// conn.WirteToUDP([]byte(daytime), addr)
					//fmt.Printf("A")
					us.dbgPrint.dbgIntervalPrint()
					mavDebug(rcv.MavlinkMessage, 1)
					//updLn.WriteToUDP(rcv.MavlinkMessage, *sendAddr)
					if us.Pair != nil {
						us.Pair.send(rcv.MavlinkMessage)
					}

				} else if err == io.EOF {
					log.Printf("grpc receiveLoop EOF e=%v", err)
					break steamLoop

				} else {
					log.Printf("grpc receiveLoop e=%v", err)
					break steamLoop
				}

			}

		} else {
			log.Print("wait OpenCommunication")
		}
		time.Sleep(10 * time.Second)
	}
}

func (us *GRPClient) send(data []byte) {
	if us.stream != nil {
		telemetory := pb.TrjxMavlink{
			MavlinkMessage: data,
		}
		if err := us.stream.Send(&telemetory); err == nil {
			//log.Printf("send error=%v", err)

		} else if err == io.EOF {
			// サーバからの切断は認証エラーおよびサーバーダウン
			tj.Logout()
			log.Printf("grpc receive EOF [%v]", err)
			st, ok := status.FromError(err)
			if ok {
				if st.Code() == codes.Unauthenticated {
					log.Printf("Unauthenticated")
				} else {
					log.Printf("grpc Code=%v", st.Code())
				}
				return
			}
			us.stream.CloseSend()
			us.stream.Context().Done()
			//us.stream = nil

		} else {
			log.Printf("grpc send error=%v", err) //SendMsg called after CloseSend：エラーでもus.stream.Recv()はEOFにならない
			time.Sleep(1 * time.Second)
			//us.stream = nil
		}
	}
}
