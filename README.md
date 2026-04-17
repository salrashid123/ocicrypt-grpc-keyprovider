## Sample gRPC server for ocicrypt

This is a simple insecure/TLS/mTLS gRPC server which uses a symmetric key to wrap encryption keys for use with [ocicrypt](https://github.com/containers/ocicrypt).

* requires [ocicrypt v1.3.0+](https://github.com/containers/ocicrypt/releases/tag/v1.3.0)

gRPC server uses AES encryption to wrap/unwrap the layer encryption key and can be accessed over

* no TLS
* TLS
* mTLS

---

| Option | Description |
|:------------|-------------|
| **`-grpcport`** | port to listen on (default: `:50051`) |
| **`-key`** | AES key to encrypt/decrypt with (default: `N1PCdw3M2B1TfJhoaY2mL736p2vCUc47`) |
| **`-rootCA`** | RootCA in PEM format to validate client certificates (default: `/certs/root-ca.crt`) |
| **`-tlsCert`** | Server TLS certificate in PEM format (default: `/certs/localhost.crt`) |
| **`-tlsKey`** | Server TLS key in PEM format (default: `/certs/localhost.key`) |
| **`-useInsecure`** | Disable TLS (default: `false`) |
| **`-usemTLS`** | Enable mTLS (default: `false`) |

---

### no TLS

```bash
go run main.go --grpcport=:50051 --useInsecure
```

The corresponding ocicrypt config would be

```json
{
  "key-providers": {
    "grpc-keyprovider": {
      "grpc": "localhost:50051"
    }
  }
}
```

### TLS

```bash
go run main.go --grpcport=:50051
```

The corresponding ocicrypt config would be

```json
{
  "key-providers": {
    "grpc-keyprovider": {
      "grpc": "localhost:50051",
      "grpc-tls": {
        "server-name": "localhost",
        "insecure-skip-verify": false,
        "root-ca-file": "/path/to/certs/root-ca.crt"
      }
    }
  }
}
```

### mTLS

```bash
go run main.go --grpcport=:50051
```

The corresponding ocicrypt config would be

```json
{
  "key-providers": {
    "grpc-keyprovider": {
      "grpc": "localhost:50051",
      "grpc-tls": {
        "server-name": "localhost",
        "insecure-skip-verify": false,
        "cert-file": "/path/to/certs/client.crt",
        "key-file":"/path/to/certs/client.key",
        "root-ca-file": "/path/to/certs/root-ca.crt"
      }
    }
  }
}
```

---

#### Imgcrypt

If you want to use [imgcrypt](https://github.com/containerd/imgcrypt) with gRPC TLS and this server and `containerd`, the configuration for `containerd` would reference a stream processor for `ctd-decoder` and the appropriate `OCICRYPT_KEYPROVIDER_CONFIG` env variable


- `config.toml`

```yaml
root = "/tmp/var/lib/containerd"
state = "/run/containerd"
temp = ""
version = 2

[debug]
  address = ""
  format = ""
  gid = 0
  level = "trace"
  uid = 0

[grpc]
  address = "/run/containerd/containerd.sock"

[plugins]

  [plugins."io.containerd.grpc.v1.cri"]
    [plugins."io.containerd.grpc.v1.cri".image_decryption]
      key_model = "node"
[stream_processors]
  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar.gzip"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+gzip+encrypted"]
    returns = "application/vnd.oci.image.layer.v1.tar+gzip"
    path = "/usr/local/bin/ctd-decoder"
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/path/to/ocicrypt.json"]
       
  [stream_processors."io.containerd.ocicrypt.decoder.v1.tar"]
    accepts = ["application/vnd.oci.image.layer.v1.tar+encrypted"]
    returns = "application/vnd.oci.image.layer.v1.tar"
    path = "/usr/local/bin/ctd-decoder"
    env = ["OCICRYPT_KEYPROVIDER_CONFIG=/path/to/ocicrypt.json"]
```

---

#### Appendix


##### Encrypted Dockerhub Image

You can find an  encrypted `busybox` image using the default AES key from this repo for testing on dockerhub

```json
$ skopeo inspect docker://docker.io/salrashid123/busybox:encrypted
{
    "Name": "docker.io/salrashid123/busybox",
    "Digest": "sha256:c7469373314a769540dd123dd1516f30be2145f143ae055523d7117d611ea4ad",
    "RepoTags": [
        "encrypted"
    ],
    "Created": "2024-09-26T21:31:42Z",
    "DockerVersion": "",
    "Labels": null,
    "Architecture": "amd64",
    "Os": "linux",
    "Layers": [
        "sha256:e19f22973137fb6598a260c3fde98df957332ec8447366744db93b30b2982e80"
    ],
    "LayersData": [
        {
            "MIMEType": "application/vnd.oci.image.layer.v1.tar+gzip+encrypted",
            "Digest": "sha256:e19f22973137fb6598a260c3fde98df957332ec8447366744db93b30b2982e80",
            "Size": 2211398,
            "Annotations": {
                "org.opencontainers.image.enc.keys.provider.grpc-keyprovider": "eyJrZXlfdXJsIjoiYWVza2V5cHJvdmlkZXI6Ly9rZXkiLCJ3cmFwcGVkX2tleSI6IlhSRXJUM1Z3UUZqNlJ6dmxFMlZxSWo3ZWE5OGxZR29jVjFXT2lOaTJjSmlzaXJkYmtFMHBBeWt1cnpERTU4MERZc2xHSUMwdG85MTgybEJXMk5HRGE0aEQyYlNCTzNsYzdsUW0yaG1pWnZaSnMzZ0RRUkUvS3NDN1Z0VGV3WlEyQnNjZFBCMytTWFFvdGVBbE9TbStqYlhRYVFkV290V0tZNWcyeXhGOERNVjlhM25mdkhGQVdtTy8wbVU0Uis5VmpiN2pDY2pTZHR1NEUrNkVoczJKK3Rkck95b3UxSHM3V0dZZUFBY2pjeEpqa2cybWtZTHl5YXR1Tk53aDBOSTdoQjRmekRFcW5JM3BnSVdFd09Ga2FXbkRpRkphSThDbkFtZVhkQUk9Iiwid3JhcF90eXBlIjoiQUVTIn0=",
                "org.opencontainers.image.enc.pubopts": "eyJjaXBoZXIiOiJBRVNfMjU2X0NUUl9ITUFDX1NIQTI1NiIsImhtYWMiOiJnTHdNNUtzc1lDR3JMKy9EdDRmZHNMbmFxR1gyd0pvNW5OcDRvc1pUbFNzPSIsImNpcGhlcm9wdGlvbnMiOnt9fQ=="
            }
        }
    ],
    "Env": [
        "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ]
}
```

```json
$ skopeo inspect docker://docker.io/salrashid123/busybox:decrypted
{
    "Name": "docker.io/salrashid123/busybox",
    "Digest": "sha256:b1347611e8b1a608d866d06d2c2f759df2a32f5e97c2b63e5bf366f7507b7244",
    "RepoTags": [
        "decrypted",
        "encrypted"
    ],
    "Created": "2024-09-26T21:31:42Z",
    "DockerVersion": "",
    "Labels": null,
    "Architecture": "amd64",
    "Os": "linux",
    "Layers": [
        "sha256:481282afbc4304ffee4792258ea114f09e423a4a082335b30695b50310394f47"
    ],
    "LayersData": [
        {
            "MIMEType": "application/vnd.oci.image.layer.v1.tar+gzip",
            "Digest": "sha256:481282afbc4304ffee4792258ea114f09e423a4a082335b30695b50310394f47",
            "Size": 2211398,
            "Annotations": null
        }
    ],
    "Env": [
        "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ]
}
```

#### Local Registry

If you want to run a local container registry to push/pull with, you can reuse the certificates here as well:

```bash
docker run  -p 5000:5000 -v `pwd`/certs:/certs \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/localhost.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/localhost.key  docker.io/registry:2

export OCICRYPT_KEYPROVIDER_CONFIG=`pwd`/ocicrypt.json
export SSL_CERT_FILE=`pwd`/certs/root-ca.crt

skopeo copy --encrypt-layer -1 \
  --encryption-key=provider:grpc-keyprovider:aeskeyprovider://key \
   docker://docker.io/busybox docker://localhost:5000/busybox:encrypted

skopeo copy \
  --decryption-key=provider:grpc-keyprovider:aeskeyprovider://key \
   docker://localhost:5000/busybox:encrypted docker://localhost:5000/busybox:decrypted

```

Where  `ocicrypt.json` may look like

```json
{
  "key-providers": {
    "grpc-keyprovider": {
      "grpc": "localhost:50051",
      "grpc-tls": {
        "server-name": "localhost",
        "insecure-skip-verify": false,
        "cert-file": "/path/to/certs/client.crt",
        "key-file":"/path/to/certs/client.key",
        "root-ca-file": "/path/to/certs/root-ca.crt"
      }
    }
  }
}
```
#### Custom Test CA

You can use any test CA to create the client and server certificates.  Here is a sample ca you can set the correct SAN values [ca_scratchpad](https://github.com/salrashid123/ca_scratchpad)

##### Cloud Run

You can run the sample grpc server locally or directly invoke it on cloud run at:

* `ocicryptgrpc-995081019036.us-central1.run.app:443`


```bash
export OCICRYPT_KEYPROVIDER_CONFIG=`pwd`/ocicrypt_cloud_run.json
export SSL_CERT_FILE=`pwd`/certs/root-ca.crt
```

- `ocicrypt_cloud_run.json`

```json
{
  "key-providers": {
    "grpc-keyprovider": {
      "grpc": "ocicryptgrpc-995081019036.us-central1.run.app:443",
      "grpc-tls": {
        "server-name": "ocicryptgrpc-995081019036.us-central1.run.app",
        "insecure-skip-verify": false,
        "root-ca-file": "/etc/ssl/certs/ca-certificates.crt"
      }
    }
  }
}
```
