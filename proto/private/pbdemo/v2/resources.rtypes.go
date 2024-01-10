// Code generated by protoc-gen-resource-types. DO NOT EDIT.

package demov2

import (
	"github.com/hashicorp/consul/proto-public/pbresource"
)

const (
	GroupName = "demo"
	Version   = "v2"
)

/* ---------------------------------------------------------------------------
 * hashicorp.consul.demo.v2.Album
 *
 * This following section contains constants variables and utility methods
 * for interacting with this kind of resource.
 * -------------------------------------------------------------------------*/
const (
	AlbumKind  = "Album"
	AlbumScope = pbresource.Scope_SCOPE_NAMESPACE
)

var AlbumType = &pbresource.Type{
	Group:        GroupName,
	GroupVersion: Version,
	Kind:         AlbumKind,
}

func (_ *Album) GetResourceType() *pbresource.Type {
	return AlbumType
}

func (_ *Album) GetResourceScope() pbresource.Scope {
	return AlbumScope
}

/* ---------------------------------------------------------------------------
 * hashicorp.consul.demo.v2.Artist
 *
 * This following section contains constants variables and utility methods
 * for interacting with this kind of resource.
 * -------------------------------------------------------------------------*/
const (
	ArtistKind  = "Artist"
	ArtistScope = pbresource.Scope_SCOPE_NAMESPACE
)

var ArtistType = &pbresource.Type{
	Group:        GroupName,
	GroupVersion: Version,
	Kind:         ArtistKind,
}

func (_ *Artist) GetResourceType() *pbresource.Type {
	return ArtistType
}

func (_ *Artist) GetResourceScope() pbresource.Scope {
	return ArtistScope
}
