.PHONY: all

all: ocicrypt_grpc_keyprovider

ocicrypt_grpc_keyprovider:
	go build -o $@ .

clean:
	rm -f ocicrypt_grpc_keyprovider
