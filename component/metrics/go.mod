module github.com/dawnsgo/dawn/component/metrics

go 1.24.0

require (
	github.com/dawnsgo/dawn v0.0.0
	github.com/dawnsgo/dawn/observe/metrics/prometheus v0.0.0
	github.com/prometheus/client_golang v1.20.5
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.2 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.11.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/dawnsgo/dawn => ../../
	github.com/dawnsgo/dawn/observe/metrics/prometheus => ../../observe/metrics/prometheus
)
