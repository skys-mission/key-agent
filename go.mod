module github.com/skys-mission/key-agent

go 1.25.0

require (
	github.com/modelcontextprotocol/go-sdk v1.4.1
	github.com/skys-mission/key-agent/keysdk v0.0.0
	github.com/spf13/cobra v1.8.1
	github.com/zalando/go-keyring v0.2.6
	go.etcd.io/bbolt v1.4.3
	golang.org/x/crypto v0.49.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/skys-mission/key-agent/keysdk => ./keysdk

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
