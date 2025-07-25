// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package ssoauth

import (
	"fmt"

	"github.com/dhiaayachi/consul/acl"
	"github.com/dhiaayachi/consul/internal/go-sso/oidcauth"
)

func validateType(typ string) error {
	if typ != "jwt" {
		return fmt.Errorf("type should be %q", "jwt")
	}
	return nil
}

func (v *Validator) ssoEntMetaFromClaims(_ *oidcauth.Claims) *acl.EnterpriseMeta {
	return nil
}

type enterpriseConfig struct{}

func (c *Config) enterpriseConvertForLibrary(_ *oidcauth.Config) {}
