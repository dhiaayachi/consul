module github.com/dhiaayachi/consul/api

go 1.23.8

replace github.com/dhiaayachi/consul/sdk => ../sdk

retract (
	v1.31.1 // checksum mismatch with tag
	v1.29.5 // cut from incorrect branch
	v1.28.0 // tag was mutated
	v1.27.1 // tag was mutated
	v1.21.2 // tag was mutated
)

require (
	github.com/dhiaayachi/consul/sdk v0.16.2
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-rootcerts v1.0.2
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/serf v0.10.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/stretchr/testify v1.8.4
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
)

require (
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/memberlist v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
