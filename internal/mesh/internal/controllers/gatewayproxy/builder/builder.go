// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/envoyextensions/xdscommon"
	"github.com/hashicorp/consul/internal/mesh/internal/controllers/gatewayproxy/fetcher"
	"github.com/hashicorp/consul/internal/mesh/internal/types"
	"github.com/hashicorp/consul/internal/resource"
	pbauth "github.com/hashicorp/consul/proto-public/pbauth/v2beta1"
	pbcatalog "github.com/hashicorp/consul/proto-public/pbcatalog/v2beta1"
	meshv2beta1 "github.com/hashicorp/consul/proto-public/pbmesh/v2beta1"
	"github.com/hashicorp/consul/proto-public/pbmesh/v2beta1/pbproxystate"
	pbmulticluster "github.com/hashicorp/consul/proto-public/pbmulticluster/v2beta1"
	"github.com/hashicorp/consul/proto-public/pbresource"
)

type proxyStateTemplateBuilder struct {
	dataFetcher      *fetcher.Fetcher
	dc               string
	exportedServices *types.DecodedComputedExportedServices
	logger           hclog.Logger
	trustDomain      string
	workload         *types.DecodedWorkload
}

func NewProxyStateTemplateBuilder(workload *types.DecodedWorkload, exportedServices *types.DecodedComputedExportedServices, logger hclog.Logger, dataFetcher *fetcher.Fetcher, dc, trustDomain string) *proxyStateTemplateBuilder {
	return &proxyStateTemplateBuilder{
		dataFetcher:      dataFetcher,
		dc:               dc,
		exportedServices: exportedServices,
		logger:           logger,
		trustDomain:      trustDomain,
		workload:         workload,
	}
}

func (b *proxyStateTemplateBuilder) identity() *pbresource.Reference {
	return &pbresource.Reference{
		Name:    b.workload.Data.Identity,
		Tenancy: b.workload.Id.Tenancy,
		Type:    pbauth.WorkloadIdentityType,
	}
}

func (b *proxyStateTemplateBuilder) listeners() []*pbproxystate.Listener {
	var address string
	if len(b.workload.Data.Addresses) > 0 {
		address = b.workload.Data.Addresses[0].Host
	}

	listener := &pbproxystate.Listener{
		Name:      xdscommon.PublicListenerName,
		Direction: pbproxystate.Direction_DIRECTION_INBOUND,
		BindAddress: &pbproxystate.Listener_HostPort{
			HostPort: &pbproxystate.HostPortAddress{
				Host: address,
				Port: b.workload.Data.Ports["wan"].Port,
			},
		},
		Capabilities: []pbproxystate.Capability{
			pbproxystate.Capability_CAPABILITY_L4_TLS_INSPECTION,
		},
		DefaultRouter: &pbproxystate.Router{
			Destination: &pbproxystate.Router_L4{
				L4: &pbproxystate.L4Destination{
					Destination: &pbproxystate.L4Destination_Cluster{
						Cluster: &pbproxystate.DestinationCluster{
							Name: "",
						},
					},
					StatPrefix: "prefix",
				},
			},
		},
		Routers: b.routers(),
	}

	// TODO NET-6429
	return []*pbproxystate.Listener{listener}
}

// routers loops through the ports and consumers for each exported service and generates
// a pbproxystate.Router matching the SNI to the target cluster. The target port name
// will be included in the ALPN. The targeted cluster will marry this port name with the SNI.
func (b *proxyStateTemplateBuilder) routers() []*pbproxystate.Router {
	var routers []*pbproxystate.Router

	for _, exportedService := range b.exportedServices.Data.Consumers {
		serviceID := resource.IDFromReference(exportedService.TargetRef)
		service, err := b.dataFetcher.FetchService(context.Background(), serviceID)
		if err != nil {
			b.logger.Trace("error reading exported service", "error", err)
			continue
		} else if service == nil {
			b.logger.Trace("service does not exist, skipping router", "service", serviceID)
			continue
		}

		for _, port := range service.Data.Ports {
			for _, consumer := range exportedService.Consumers {
				routers = append(routers, &pbproxystate.Router{
					Match: &pbproxystate.Match{
						AlpnProtocols: []string{alpnProtocol(port.TargetPort)},
						ServerNames:   []string{b.sni(exportedService.TargetRef, consumer)},
					},
					Destination: &pbproxystate.Router_L4{
						L4: &pbproxystate.L4Destination{
							Destination: &pbproxystate.L4Destination_Cluster{
								Cluster: &pbproxystate.DestinationCluster{
									Name: b.clusterName(exportedService.TargetRef, consumer, port.TargetPort),
								},
							},
							StatPrefix: "prefix",
						},
					},
				})
			}
		}
	}

	return routers
}

// clusters loops through the consumers for each exported service
// and generates a pbproxystate.Cluster per service-consumer pairing.
func (b *proxyStateTemplateBuilder) clusters() map[string]*pbproxystate.Cluster {
	clusters := map[string]*pbproxystate.Cluster{}

	for _, exportedService := range b.exportedServices.Data.Consumers {
		serviceID := resource.IDFromReference(exportedService.TargetRef)
		service, err := b.dataFetcher.FetchService(context.Background(), serviceID)
		if err != nil {
			b.logger.Trace("error reading exported service", "error", err)
			continue
		} else if service == nil {
			b.logger.Trace("service does not exist, skipping router", "service", serviceID)
			continue
		}

		for _, port := range service.Data.Ports {
			for _, consumer := range exportedService.Consumers {
				clusterName := b.clusterName(exportedService.TargetRef, consumer, port.TargetPort)
				clusters[clusterName] = &pbproxystate.Cluster{
					Name:     clusterName,
					Protocol: pbproxystate.Protocol_PROTOCOL_TCP, // TODO
					Group: &pbproxystate.Cluster_EndpointGroup{
						EndpointGroup: &pbproxystate.EndpointGroup{
							Group: &pbproxystate.EndpointGroup_Dynamic{},
						},
					},
					AltStatName: "prefix",
				}
			}
		}
	}

	return clusters
}

// requiredEndpoints loops through the consumers for each exported service
// and adds a pbproxystate.EndpointRef to be hydrated for each cluster.
func (b *proxyStateTemplateBuilder) requiredEndpoints() map[string]*pbproxystate.EndpointRef {
	requiredEndpoints := map[string]*pbproxystate.EndpointRef{}

	for _, exportedService := range b.exportedServices.Data.Consumers {
		serviceID := resource.IDFromReference(exportedService.TargetRef)
		service, err := b.dataFetcher.FetchService(context.Background(), serviceID)
		if err != nil {
			b.logger.Trace("error reading exported service", "error", err)
			continue
		} else if service == nil {
			b.logger.Trace("service does not exist, skipping router", "service", serviceID)
			continue
		}

		for _, port := range service.Data.Ports {
			for _, consumer := range exportedService.Consumers {
				clusterName := b.clusterName(exportedService.TargetRef, consumer, port.TargetPort)
				requiredEndpoints[clusterName] = &pbproxystate.EndpointRef{
					Id:   resource.ReplaceType(pbcatalog.ServiceEndpointsType, serviceID),
					Port: port.TargetPort,
				}
			}
		}
	}

	return requiredEndpoints
}

func (b *proxyStateTemplateBuilder) endpoints() map[string]*pbproxystate.Endpoints {
	// TODO NET-6431
	return nil
}

func (b *proxyStateTemplateBuilder) routes() map[string]*pbproxystate.Route {
	// TODO NET-6428
	return nil
}

func (b *proxyStateTemplateBuilder) Build() *meshv2beta1.ProxyStateTemplate {
	return &meshv2beta1.ProxyStateTemplate{
		ProxyState: &meshv2beta1.ProxyState{
			Identity:  b.identity(),
			Listeners: b.listeners(),
			Clusters:  b.clusters(),
			Endpoints: b.endpoints(),
			Routes:    b.routes(),
		},
		RequiredEndpoints:        b.requiredEndpoints(),
		RequiredLeafCertificates: make(map[string]*pbproxystate.LeafCertificateRef),
		RequiredTrustBundles:     make(map[string]*pbproxystate.TrustBundleRef),
	}
}

func (b *proxyStateTemplateBuilder) clusterName(serviceRef *pbresource.Reference, consumer *pbmulticluster.ComputedExportedServicesConsumer, port string) string {
	return fmt.Sprintf("%s.%s", port, b.sni(serviceRef, consumer))
}

func (b *proxyStateTemplateBuilder) sni(serviceRef *pbresource.Reference, consumer *pbmulticluster.ComputedExportedServicesConsumer) string {
	switch tConsumer := consumer.ConsumerTenancy.(type) {
	case *pbmulticluster.ComputedExportedServicesConsumer_Partition:
		return connect.ServiceSNI(serviceRef.Name, "", serviceRef.Tenancy.Namespace, tConsumer.Partition, b.dc, b.trustDomain)
	case *pbmulticluster.ComputedExportedServicesConsumer_Peer:
		return connect.PeeredServiceSNI(serviceRef.Name, serviceRef.Tenancy.Namespace, serviceRef.Tenancy.Partition, tConsumer.Peer, b.trustDomain)
	default:
		return ""
	}
}

func alpnProtocol(portName string) string {
	return fmt.Sprintf("consul~%s", portName)
}