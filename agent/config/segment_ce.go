// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package config

import (
	"github.com/dhiaayachi/consul/agent/structs"
)

func (b *builder) validateSegments(rt RuntimeConfig) error {
	if rt.SegmentName != "" {
		return structs.ErrSegmentsNotSupported
	}
	if len(rt.Segments) > 0 {
		return structs.ErrSegmentsNotSupported
	}
	return nil
}
