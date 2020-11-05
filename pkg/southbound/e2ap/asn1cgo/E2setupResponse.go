// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package asn1cgo

//#cgo CFLAGS: -I. -D_DEFAULT_SOURCE -DASN_DISABLE_OER_SUPPORT
//#cgo LDFLAGS: -lm
//#include <stdio.h>
//#include <stdlib.h>
//#include <assert.h>
//#include "E2setupResponse.h"
//#include "ProtocolIE-Field.h"
import "C"
import (
	"github.com/onosproject/onos-e2t/api/e2ap/v1beta1/e2appducontents"
	"unsafe"
)

func xerEncodeE2setupResponse(e2sr *e2appducontents.E2SetupResponse) ([]byte, error) {
	e2SetupResponseC, err := newE2setupResponse(e2sr)
	if err != nil {
		return nil, err
	}

	bytes, err := encodeXer(&C.asn_DEF_E2setupResponse, unsafe.Pointer(e2SetupResponseC))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func perEncodeE2setupResponse(e2sr *e2appducontents.E2SetupResponse) ([]byte, error) {
	e2SetupResponseC, err := newE2setupResponse(e2sr)
	if err != nil {
		return nil, err
	}

	bytes, err := encodePerBuffer(&C.asn_DEF_E2setupResponse, unsafe.Pointer(e2SetupResponseC))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func newE2setupResponse(e2sr *e2appducontents.E2SetupResponse) (*C.E2setupResponse_t, error) {
	pIeC1544P12, err := newE2SetupResponseIes(e2sr.ProtocolIes)
	if err != nil {
		return nil, err
	}
	rsrC := C.E2setupResponse_t{
		protocolIEs: *pIeC1544P12,
	}

	return &rsrC, nil
}