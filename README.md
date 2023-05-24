# dippy

Simple TCP Proxy

## Usage

```
dippy -config=0.0.0.0:80=192.168.0.2:8080,0.0.0.0:443=192.168.0.3:8443

# or with env
PROXY_CONFIG=0.0.0.0:80=192.168.0.2:8080,0.0.0.0:443=192.168.0.3:8443 dippy
```

## License

MIT
