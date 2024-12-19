package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/chyiyaqing/gmicro-payment/config"
	"github.com/chyiyaqing/gmicro-payment/internal/ports"
	"github.com/chyiyaqing/gmicro-proto/golang/payment"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *grpc.Server
	payment.UnimplementedPaymentServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

func getTlsCredentials() (credentials.TransportCredentials, error) {
	serverCert, serverCertErr := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem") // 加载服务器证书
	// handle serverCertErr
	if serverCertErr != nil {
		return nil, serverCertErr
	}
	certPool := x509.NewCertPool() // CA检查的证书池子
	caCert, caCertErr := os.ReadFile("cert/ca-cert.pem")
	if caCertErr != nil {
		return nil, caCertErr
	}

	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("failed to append the CA certs")
	}
	return credentials.NewTLS(
		&tls.Config{
			ClientAuth:   tls.RequireAnyClientCert,
			Certificates: []tls.Certificate{serverCert},
			ClientCAs:    certPool,
		}), nil
}

func (a *Adapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}
	tlsCredentials, tlsCredentialsErr := getTlsCredentials()
	if tlsCredentialsErr != nil {
		log.Fatalf("failed to get tls credential, error: %v", tlsCredentialsErr)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.Creds(tlsCredentials), grpc.StatsHandler(otelgrpc.NewServerHandler()))
	grpcServer := grpc.NewServer(opts...)
	a.server = grpcServer
	payment.RegisterPaymentServer(grpcServer, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}

	log.Printf("starting payment service on port %d ...", a.port)
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port")
	}
}

func (a *Adapter) Stop() {
	a.server.Stop()
}
