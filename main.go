package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/containers/ocicrypt/keywrap/keyprovider"
	keyproviderpb "github.com/containers/ocicrypt/utils/keyprovider"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	grpcport    = flag.String("grpcport", ":50051", "grpcport")
	useInsecure = flag.Bool("useInsecure", false, "do not use TLS")
	usemTLS     = flag.Bool("usemTLS", false, "use mtls for grpc server")
	tlsCert     = flag.String("tlsCert", "certs/localhost.crt", "tls Certificate")
	tlsKey      = flag.String("tlsKey", "certs/localhost.key", "tls Key")
	rootCA      = flag.String("rootCA", "certs/root-ca.crt", "CA")
	key         = flag.String("key", "N1PCdw3M2B1TfJhoaY2mL736p2vCUc47", "Key")
)

const (
	grpcProvider = "grpc-keyprovider"
	providerURI  = "aeskeyprovider://key"
)

type server struct {
	keyproviderpb.UnimplementedKeyProviderServiceServer
}

type annotationPacket struct {
	KeyUrl     string `json:"key_url"`
	WrappedKey []byte `json:"wrapped_key"`
	WrapType   string `json:"wrap_type"`
}

func (*server) WrapKey(ctx context.Context, request *keyproviderpb.KeyProviderKeyWrapProtocolInput) (*keyproviderpb.KeyProviderKeyWrapProtocolOutput, error) {
	var keyP keyprovider.KeyProviderKeyWrapProtocolInput
	err := json.Unmarshal(request.KeyProviderKeyWrapProtocolInput, &keyP)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(*key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, keyP.KeyWrapParams.OptsData, nil)

	jsonString, err := json.Marshal(annotationPacket{
		KeyUrl:     providerURI,
		WrappedKey: ciphertext,
		WrapType:   "AES",
	})
	if err != nil {
		return nil, err
	}

	protocolOuputSerialized, err := json.Marshal(keyprovider.KeyProviderKeyWrapProtocolOutput{
		KeyWrapResults: keyprovider.KeyWrapResults{Annotation: jsonString},
	})
	if err != nil {
		return nil, fmt.Errorf("Error marshalling KeyProviderKeyWrapProtocolOutput %v\n", err)
	}

	return &keyproviderpb.KeyProviderKeyWrapProtocolOutput{
		KeyProviderKeyWrapProtocolOutput: protocolOuputSerialized,
	}, nil

}

func (*server) UnWrapKey(ctx context.Context, request *keyproviderpb.KeyProviderKeyWrapProtocolInput) (*keyproviderpb.KeyProviderKeyWrapProtocolOutput, error) {

	var keyP keyprovider.KeyProviderKeyWrapProtocolInput
	err := json.Unmarshal(request.KeyProviderKeyWrapProtocolInput, &keyP)
	if err != nil {
		return nil, err
	}
	apkt := annotationPacket{}
	err = json.Unmarshal(keyP.KeyUnwrapParams.Annotation, &apkt)
	if err != nil {
		return nil, err
	}

	ciphertext := apkt.WrappedKey

	block, err := aes.NewCipher([]byte(*key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, err
	}

	protocolOuputSerialized, err := json.Marshal(keyprovider.KeyProviderKeyWrapProtocolOutput{
		KeyUnwrapResults: keyprovider.KeyUnwrapResults{OptsData: plaintext},
	})
	if err != nil {
		return nil, fmt.Errorf("Error marshalling KeyProviderKeyWrapProtocolOutput %v\n", err)
	}

	return &keyproviderpb.KeyProviderKeyWrapProtocolOutput{
		KeyProviderKeyWrapProtocolOutput: protocolOuputSerialized,
	}, nil
}

func main() {

	flag.Parse()

	var err error

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	if !*useInsecure {
		clientCaCert, err := os.ReadFile(*rootCA)
		if err != nil {
			log.Fatalf("mtls enabled but cannot read mtlsBackendCA cert")
		}
		clientCaCertPool := x509.NewCertPool()
		clientCaCertPool.AppendCertsFromPEM(clientCaCert)

		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("failed to load client certificate and key: %v", err)
		}

		var tlsConfig *tls.Config

		if *usemTLS {
			tlsConfig = &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    clientCaCertPool,
				Certificates: []tls.Certificate{cert},
			}
		} else {
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		}

		creds := credentials.NewTLS(tlsConfig)

		sopts = append(sopts, grpc.Creds(creds))
	} else {
		sopts = append(sopts, grpc.Creds(insecure.NewCredentials()))
	}

	s := grpc.NewServer(sopts...)
	keyproviderpb.RegisterKeyProviderServiceServer(s, &server{})

	log.Printf("Starting gRPC Server at %s", *grpcport)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-done
		log.Printf("caught sig: %+v", sig)
		log.Println("Wait for 1 second to finish processing")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Error starting server %v", err)
	}

}
