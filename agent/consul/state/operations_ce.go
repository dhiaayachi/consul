// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build !consulent

package state

import (
	"github.com/hashicorp/go-memdb"

	"github.com/dhiaayachi/consul/acl"
)

func getCompoundWithTxn(tx ReadTxn, table, index string,
	_ *acl.EnterpriseMeta, idxVals ...interface{}) (memdb.ResultIterator, error) {

	return tx.Get(table, index, idxVals...)
}
