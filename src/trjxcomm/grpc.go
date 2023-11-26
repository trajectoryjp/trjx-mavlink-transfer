package trjxcomm

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "gomav/code/proto/pb"
	fi "gomav/code/trjxtransfer/fileio"
	fi "github.com/trajectoryjp/trjx-mavlink-transfer/fileio"
)

var tmService pb.TrjxMavlinkServiceClient
var connection *grpc.ClientConn
var token string

//var openStream chan struct{}

/*
// gRPCクライアントセットアップ
// 認証
// Loginが承認されるまで繰り返し

func setupRPC() {

	go watchLoop()

}
*/

func Login(aircraftID string, addressInfo fi.Address) {
	wait := 1

	for {
		// connection確立ループ

		if connection == nil || tmService == nil {
			var tc credentials.TransportCredentials
			if addressInfo.UseTLS {
				log.Printf("use TLS\n")
				c := tls.Config{ServerName: addressInfo.ServerName}
				tc = credentials.NewTLS(&c)
			} else {
				log.Printf("not use TLS\n")
				tc = insecure.NewCredentials()
			}

			//address := "127.0.0.1:50052"
			//address := fi.TrjxAircraftConfigData.TRJXServer
			address := fmt.Sprintf("%s:%d", addressInfo.Address, addressInfo.Port)
			var err error

			connection, err = grpc.Dial(address,
				grpc.WithTransportCredentials(tc),
				grpc.WithUnaryInterceptor(DoUnaryClientInterceptor),
				grpc.WithStreamInterceptor(DoStreamClientInterceptor),
			)

			if err == nil && connection != nil {
				log.Printf("gRPC successful %v\n", address)
				wait = 1

			} else {
				log.Printf("gRPC connect error=%v wait=%v[sec]", err, wait)
				if connection != nil {
					connection.Close()
					connection = nil
				}
				time.Sleep(time.Duration(wait) * time.Second)
				wait *= 2
				if wait > 300 {
					wait = 300
				}
			}
			if tmService = pb.NewTrjxMavlinkServiceClient(connection); tmService == nil {
				connection.Close()
				connection = nil
			}

		} else {

			if token == "" {
				//aircraftID := fi.TrjxAircraftConfigData.AircraftID
				password := fi.ReadPassword()

				// ログイン
				md := metadata.New(map[string]string{"aircraft": aircraftID, "password": password})
				ctx := metadata.NewOutgoingContext(context.Background(), md)
				log.Printf("Login metadata=%v\n", md)
				req := pb.Empty{}
				if response, err := tmService.Login(ctx, &req); err == nil {
					switch response.Result {
					case pb.Token_Complete, pb.Token_Accepted:
						token = response.Token
						password := response.Password
						if response.Password == "" {
							log.Fatal("passowrd in null")
						}
						// passwordセーブ
						log.Printf("password=%v token=%v", password, token)
						fi.WritePassword(password)

					default:
						Logout()
						log.Printf("Error gRPC login code=%v\n", response.Result)
						time.Sleep(time.Duration(wait) * time.Second)
					}

				} else {
					Logout()
					log.Printf("Error gRPC login wait=%v e==%v\n", wait, err)
					tmService = nil
					time.Sleep(time.Duration(wait) * time.Second)
				}
				wait *= 2
				if wait > 300 {
					wait = 300
				}

			} else {
				return
			}

		}
	}
	log.Printf("### Login-end\n")
}

func Logout() {
	token = ""
}

func OpenCommunication(aircraft string, address fi.Address) pb.TrjxMavlinkService_CommunicateOnMavlinkClient {
	// 認証処理
	for {
		if tmService != nil && token != "" {
			// https://zenn.dev/miyazi777/articles/a560e691fcee0b6449e4
			//aircraft := fi.TrjxAircraftConfigData.AircraftID
			log.Printf("try CommunicateOnMavlink gRPC")

			md := metadata.New(map[string]string{"aircraft": aircraft, "token": token})
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			log.Printf("OpenCommunication metadata=%v\n", md)
			if stream, err := tmService.CommunicateOnMavlink(ctx); err == nil {
				if stream != nil {
					log.Printf("connected CommunicateOnMavlink gRPC")
					return stream
				}
				log.Printf("openCommunication CommunicateOnMavlink stream is nil")

			} else {
				log.Printf("openCommunication CommunicateOnMavlink err=%v stream=%v", err, stream)
			}
		}
		tmService = nil
		if connection != nil {
			log.Printf("connection close")
			connection.Close()
			connection = nil
		}
		log.Printf("try Login")
		Login(aircraft, address)
		time.Sleep(10 * time.Second)
	}

}

func SetMetadata(ctx context.Context, method string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	log.Printf("SetMetadata ok=%v metadata=%v\n", ok, md)
	if ok {
		if md != nil {
			return metadata.AppendToOutgoingContext(ctx, "x-grpc-service", "trjx-transfer")
		}
	}
	md = metadata.New(map[string]string{"x-grpc-service": "trjx-transfer"})
	return metadata.NewOutgoingContext(ctx, md)
}

type ClientStreamWrapper struct {
	grpc.ClientStream
}

func DoUnaryClientInterceptor(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = SetMetadata(ctx, method)
	md, ok := metadata.FromOutgoingContext(ctx)
	log.Printf("DoUnaryClientInterceptor ok=%v metadata=%v\n", ok, md)
	err := invoker(ctx, method, req, res, cc, opts...)
	return err
}

func DoStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = SetMetadata(ctx, method)
	md, ok := metadata.FromOutgoingContext(ctx)
	log.Printf("DoStreamClientInterceptor ok=%v metadata=%v\n", ok, md)
	stream, err := streamer(ctx, desc, cc, method, opts...)
	return &ClientStreamWrapper{stream}, err
}
