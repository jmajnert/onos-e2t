// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package channels

import (
	"context"
	"io"
	"net"

	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v2beta1/e2ap-pdu-contents"
	e2appdudescriptions "github.com/onosproject/onos-e2t/api/e2ap/v2beta1/e2ap-pdu-descriptions"
	"github.com/onosproject/onos-e2t/pkg/protocols/e2ap201/procedures"
	"github.com/onosproject/onos-e2t/pkg/utils/async"
)

// RICHandler is a function for wrapping an RICChannel
type RICHandler func(channel RICChannel) procedures.RICProcedures

// RICChannel is a channel for an E2 node
type RICChannel interface {
	Channel
	procedures.E2NodeProcedures
}

// NewRICChannel creates a new E2 node channel
func NewRICChannel(conn net.Conn, handler RICHandler, opts ...Option) RICChannel {
	parent := newThreadSafeChannel(conn, opts...)
	channel := &ricChannel{
		threadSafeChannel: parent,
	}
	procs := handler(channel)
	channel.e2Setup = procedures.NewE2SetupProcedure(parent.send, procs)
	channel.e2ConfigurationUpdate = procedures.NewE2ConfigurationUpdateProcedure(parent.send, procs)
	channel.e2ConnectionUpdate = procedures.NewE2ConnectionUpdateInitiator(parent.send)
	channel.ricControl = procedures.NewRICControlInitiator(parent.send)
	channel.ricIndication = procedures.NewRICIndicationProcedure(parent.send, procs)
	channel.ricSubscription = procedures.NewRICSubscriptionInitiator(parent.send)
	channel.ricSubscriptionDelete = procedures.NewRICSubscriptionDeleteInitiator(parent.send)
	channel.open()
	return channel
}

// ricChannel is an E2 node channel
type ricChannel struct {
	*threadSafeChannel
	e2Setup               *procedures.E2SetupProcedure
	e2ConfigurationUpdate *procedures.E2ConfigurationUpdateProcedure
	e2ConnectionUpdate    *procedures.E2ConnectionUpdateInitiator
	ricControl            *procedures.RICControlInitiator
	ricIndication         *procedures.RICIndicationProcedure
	ricSubscription       *procedures.RICSubscriptionInitiator
	ricSubscriptionDelete *procedures.RICSubscriptionDeleteInitiator
	ricIndicationCh       chan *e2appdudescriptions.E2ApPdu
}

func (c *ricChannel) open() {
	c.ricIndicationCh = make(chan *e2appdudescriptions.E2ApPdu)
	go c.recvPDUs()
	go c.recvIndications()
}

func (c *ricChannel) recvPDUs() {
	for {
		pdu, err := c.recv()
		if err == io.EOF {
			log.Warn(err)
			c.Close()
			return
		}
		if err != nil {
			log.Error(err)
		} else {
			c.recvPDU(pdu)
		}
	}
}

func (c *ricChannel) recvPDU(pdu *e2appdudescriptions.E2ApPdu) {
	if c.e2Setup.Matches(pdu) {
		go c.e2Setup.Handle(pdu)
	} else if c.e2ConfigurationUpdate.Matches(pdu) {
		go c.e2ConfigurationUpdate.Handle(pdu)
	} else if c.e2ConnectionUpdate.Matches(pdu) {
		go c.e2ConnectionUpdate.Handle(pdu)
	} else if c.ricControl.Matches(pdu) {
		go c.ricControl.Handle(pdu)
	} else if c.ricIndication.Matches(pdu) {
		c.ricIndicationCh <- pdu
	} else if c.ricSubscription.Matches(pdu) {
		go c.ricSubscription.Handle(pdu)
	} else if c.ricSubscriptionDelete.Matches(pdu) {
		go c.ricSubscriptionDelete.Handle(pdu)
	} else {
		log.Errorf("Unsupported E2AP message: %+v", pdu)
	}
}

func (c *ricChannel) recvIndications() {
	for pdu := range c.ricIndicationCh {
		c.recvIndication(pdu)
	}
}

func (c *ricChannel) recvIndication(pdu *e2appdudescriptions.E2ApPdu) {
	c.ricIndication.Handle(pdu)
}

func (c *ricChannel) E2ConnectionUpdate(ctx context.Context, request *e2appducontents.E2ConnectionUpdate) (response *e2appducontents.E2ConnectionUpdateAcknowledge, failure *e2appducontents.E2ConnectionUpdateFailure, err error) {
	return c.e2ConnectionUpdate.Initiate(ctx, request)
}

func (c *ricChannel) RICControl(ctx context.Context, request *e2appducontents.RiccontrolRequest) (response *e2appducontents.RiccontrolAcknowledge, failure *e2appducontents.RiccontrolFailure, err error) {
	return c.ricControl.Initiate(ctx, request)
}

func (c *ricChannel) RICSubscription(ctx context.Context, request *e2appducontents.RicsubscriptionRequest) (response *e2appducontents.RicsubscriptionResponse, failure *e2appducontents.RicsubscriptionFailure, err error) {
	return c.ricSubscription.Initiate(ctx, request)
}

func (c *ricChannel) RICSubscriptionDelete(ctx context.Context, request *e2appducontents.RicsubscriptionDeleteRequest) (response *e2appducontents.RicsubscriptionDeleteResponse, failure *e2appducontents.RicsubscriptionDeleteFailure, err error) {
	return c.ricSubscriptionDelete.Initiate(ctx, request)
}

func (c *ricChannel) Close() error {
	procedures := []procedures.ElementaryProcedure{
		c.e2Setup,
		c.e2ConfigurationUpdate,
		c.e2ConnectionUpdate,
		c.ricControl,
		c.ricIndication,
		c.ricSubscription,
		c.ricSubscriptionDelete,
	}
	err := async.Apply(len(procedures), func(i int) error {
		return procedures[i].Close()
	})
	if err != nil {
		return err
	}
	return nil
}

var _ RICChannel = &ricChannel{}