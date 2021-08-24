module github.com/hyperchain/gosdk

require (
	github.com/aristanetworks/goarista v0.0.0-20190325233358-a123909ec740
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/coreos/etcd v3.3.12+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/dsnet/compress v0.0.0-20171208185109-cc9eb1d7ad76 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.1
	github.com/gorilla/websocket v0.0.0-20180605202552-5ed622c449da
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.7
	github.com/kr/pretty v0.1.0 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/mholt/archiver v0.0.0-20180417220235-e4ef56d48eb0
	github.com/mitchellh/mapstructure v1.2.2
	github.com/nwaples/rardecode v0.0.0-20171029023500-e06696f847ae // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/opentracing/opentracing-go v0.0.0-20180606204148-bd9c31933947
	github.com/pelletier/go-toml v1.3.0 // indirect
	github.com/pierrec/lz4 v2.0.5+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streadway/amqp v0.0.0-20180528204448-e5adc2ada8b8
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/terasum/viper v0.0.0-20170802085632-7507f719f06e
	github.com/ulikunitz/xz v0.5.4 // indirect
	github.com/ultramesh/crypto-gm v0.2.8
	github.com/ultramesh/crypto-standard v0.1.12
	github.com/ultramesh/flato-msp-cert v0.1.5
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace github.com/ultramesh/crypto-standard => git.hyperchain.cn/ultramesh/crypto-standard.git v0.1.13

replace github.com/ultramesh/crypto-gm => git.hyperchain.cn/ultramesh/crypto-gm.git v0.2.8

replace github.com/ultramesh/flato-msp-cert => git.hyperchain.cn/ultramesh/flato-msp-cert.git v0.1.5

go 1.13
