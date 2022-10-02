// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package conn

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"v.io/v23/context"
	"v.io/v23/flow"
	"v.io/v23/flow/message"
	"v.io/v23/naming"
	"v.io/v23/security"
)

type flw struct {
	// These variables are all set during flow construction.
	id                                uint64
	conn                              *Conn
	q                                 *readq
	localBlessings, remoteBlessings   security.Blessings
	localDischarges, remoteDischarges map[string]security.Discharge
	noEncrypt                         bool
	noFragment                        bool
	remote                            naming.Endpoint
	channelTimeout                    time.Duration
	sideChannel                       bool
	encapsulated                      bool

	// The following fields for managing flow control and the writeq
	// are locked indepdently.
	flowControl flowControlFlowStats

	writeq      *writeq
	writeqEntry writer

	// mu guards all of the following fields.
	mu     sync.Mutex
	ctx    *context.T
	cancel context.CancelFunc

	// opened indicates whether the flow has already been opened.  If false
	// we need to send an open flow on the next write.  For accepted flows
	// this will always be true.
	opened   bool
	openSent chan struct{}

	// tokenWait is set if a write is blocked waiting for flowcontrol
	// tokens. When tokens become available the channel should be closed.
	tokenWait chan struct{}

	// closed is true as soon as f.close has been called.
	closed bool
}

func (f *flw) lock() {
	f.mu.Lock()
}

func (f *flw) unlock() {
	f.mu.Unlock()
}

// Ensure that *flw implements flow.Flow.
var _ flow.Flow = &flw{}

func (c *Conn) newFlowLocked(
	ctx *context.T,
	id uint64,
	localBlessings, remoteBlessings security.Blessings,
	localDischarges, remoteDischarges map[string]security.Discharge,
	remote naming.Endpoint,
	dialed bool,
	channelTimeout time.Duration,
	sideChannel bool) *flw {
	f := &flw{
		id:               id,
		conn:             c,
		localBlessings:   localBlessings,
		localDischarges:  localDischarges,
		remoteBlessings:  remoteBlessings,
		remoteDischarges: remoteDischarges,
		opened:           !dialed,
		remote:           remote,
		channelTimeout:   channelTimeout,
		sideChannel:      sideChannel,
		writeq:           &c.writeq,
		encapsulated:     c.IsEncapsulated(),
	}
	// It's important that this channel has a non-zero buffer since flows
	// will be notifying themselve and if there's no buffer a deadlock will
	// occur. The self notification is between the code that handles release
	// messages to notify a flow-controlled writeMgs that it may potentially
	// have tokens to spend on writes.
	initWriter(&f.writeqEntry, 1)
	if dialed {
		f.writeqEntry.msg = fmt.Sprintf("flow %p, id %v: dialed", f, id)
	} else {
		f.writeqEntry.msg = fmt.Sprintf("flow %p, id %v: accepted", f, id)
	}

	f.flowControl.shared = &c.flowControl
	f.flowControl.borrowing = dialed
	f.flowControl.id = id

	if f.opened {
		f.openSent = make(chan struct{})
		close(f.openSent)
	}
	f.tokenWait = make(chan struct{})

	f.q = newReadQ(f.flowControl.shared.bytesBufferedPerFlow, f.sendRelease)

	f.ctx, f.cancel = context.WithCancel(ctx)
	if !f.opened {
		c.unopenedFlows.Add(1)
	}
	c.flows[id] = f
	c.healthCheckNewFlowLocked(ctx, channelTimeout)
	return f
}

func (f *flw) sendRelease(ctx *context.T, n int) {
	f.conn.sendRelease(ctx, f.id, uint64(n))
}

// disableEncrytion should not be called concurrently with Write* methods.
func (f *flw) disableEncryption() {
	f.noEncrypt = true
}

// DisableFragmentation should probably not be called concurrently with
// Write* methods.
func (f *flw) DisableFragmentation() {
	f.noFragment = true
}

// Implement io.Reader.
// Read and ReadMsg should not be called concurrently with themselves
// or each other.
func (f *flw) Read(p []byte) (n int, err error) {
	ctx := f.currentContext()
	f.markUsed()
	if n, err = f.q.read(ctx, p); err != nil {
		f.close(ctx, false, err)
	}
	return
}

// ReadMsg is like read, but it reads bytes in chunks.  Depending on the
// implementation the batch boundaries might or might not be significant.
// Read and ReadMsg should not be called concurrently with themselves
// or each other.
func (f *flw) ReadMsg() (buf []byte, err error) {
	ctx := f.currentContext()
	f.markUsed()
	// TODO(mattr): Currently we only ever release counters when some flow
	// reads.  We may need to do it more or less often.  Currently
	// we'll send counters whenever a new flow is opened.
	if buf, err = f.q.get(ctx); err != nil {
		f.close(ctx, false, err)
	}
	return
}

// ReadMsg2 is like ReadMsg. In this implementation it does not use the
// supplied buffer since doing so would force an extraneous allocation and
// copy.
func (f *flw) ReadMsg2(_ []byte) (buf []byte, err error) {
	return f.ReadMsg()
}

// Implement io.Writer.
// Write, WriteMsg, and WriteMsgAndClose should not be called concurrently
// with themselves or each other.
func (f *flw) Write(p []byte) (n int, err error) {
	return f.WriteMsg(p)
}

func (f *flw) currentContext() *context.T {
	f.lock()
	defer f.unlock()
	return f.ctx
}

// releaseCounters releases some counters from a remote reader to the local
// writer.  This allows the writer to then write more data to the wire.
func (f *flw) releaseCounters(tokens uint64) {
	ctx := f.currentContext()
	f.lock()
	defer f.unlock()
	f.flowControl.releaseCounters(ctx, tokens)
	fmt.Fprintf(os.Stderr, "flow.control: %p:%p:%3v: releaseCounters: #%v tokens\n", f.conn, f, f.id, tokens)
	// Broadcast to all flows blocked waiting for tokens that some may be availabe
	// now.
	close(f.tokenWait)
	f.tokenWait = make(chan struct{})
	if ctx.V(2) {
		ctx.Infof("Activated writing flow %d(%p) now that we have tokens.", f.id, f)
	}
}

// ensureTokens ensures that there are sufficient flow control tokens available
// to send the current message, and if there are none, it will wait until
// some become available from the peer end of the connection via a release
// message containing tokens for this flow.
func (f *flw) ensureTokens(ctx *context.T, debug bool, need int) (int, func(int), error) {
	for {
		// The critical region needs to capture both reading tokens and
		// setting up to wait for new ones. Similarly, releaseCounters
		// above needs to lock access to updating the available tokens
		// (ie. the call to releaseCounters) as well as closing the channel
		// used to notify the code below that more tokens may potentially
		// be available.
		f.lock()
		avail, deduct := f.flowControl.tokens(ctx, f.encapsulated)
		if avail >= need {
			fmt.Fprintf(os.Stderr, "flow.control: ensureTokens: %p:%p:%3v: have flow control: %v:%v\n", f.conn, f, f.id, avail, need)
			f.unlock()
			return avail, deduct, nil
		}
		fmt.Fprintf(os.Stderr, "flow.control: ensureTokens: %p:%p:%3v: need flow control: %v:%v\n", f.conn, f, f.id, avail, need)
		// need to wait for tokens.
		if debug {
			ctx.Infof("Deactivating write on flow %d(%p) due to lack of tokens", f.id, f)
		}
		ch := f.tokenWait
		f.unlock()
		select {
		case <-ctx.Done():
			return 0, func(int) {}, ctx.Err()
		case <-ch:
		}
		fmt.Fprintf(os.Stderr, "flow.control: ensureTokens: %p:%p:%3v: updated flow control: %v:%v\n", f.conn, f, f.id, avail, need)
	}
}

func (f *flw) writeMsg(alsoClose bool, parts ...[]byte) (sent int, err error) {
	ctx := f.currentContext()
	select {
	// Catch cancellations early.  If we caught a cancel when waiting
	// our turn below it's possible that we were notified simultaneously.
	// Then the notify channel will be full and we would deadlock
	// notifying ourselves.
	case <-ctx.Done():
		f.close(ctx, false, ctx.Err())
		return 0, io.EOF
	default:
	}

	// TODO: don't send blessings more than once.
	bkey, dkey, err := f.conn.blessingsFlow.send(ctx, f.localBlessings, f.localDischarges, nil)
	if err != nil {
		return 0, err
	}

	debug := f.ctx.V(2)
	if debug {
		ctx.Infof("starting write on flow %d(%p)", f.id, f)
	}

	totalSize := 0
	for _, p := range parts {
		totalSize += len(p)
	}
	size, sent, tosend := 0, 0, make([][]byte, len(parts))

	f.markUsed()

	for len(parts) > 0 {

		// We're prepared to send at least 1 byte under any circumstance.
		need := 1
		if f.noEncrypt || f.noFragment {
			// Note that if f.noEncrypt is set we're actually acting as a conn
			// for higher level flows. In this case we don't want to fragment the
			// writes of the higher level flows, we want to transmit their messages
			// whole. Similarly if noFragment is set we prefer to wait and not
			// to send a partial write based on the available tokens. This is only
			// required by proxies which need to pass on messages without
			// refragmenting.
			need = totalSize
		}

		f.lock()
		wasOpened := f.setOpenedLocked()
		if !wasOpened {
			f.openSent = make(chan struct{})
			f.unlock()
		} else {
			ch := f.openSent
			f.unlock()
			<-ch
			tokens, deduct, terr := f.ensureTokens(ctx, debug, need)
			parts, tosend, size = popFront(parts, tosend[:0], tokens)
			deduct(size)
			err = terr
			fmt.Fprintf(os.Stderr, "flow: writeMsg: %p:%p:%3v: %p:%p: after flow control and before writeq: %v %v/%v (alsoclose: %v)\n", f.conn, f, f.id, f.writeq, &f.writeqEntry, tokens, sent, totalSize, alsoClose)
		}

		err = f.sendFlowMessage(ctx, wasOpened, alsoClose, len(parts) == 0, bkey, dkey, tosend)
		if err != nil {
			break
		}
		if !wasOpened {
			close(f.openSent)
		}
		sent += size

		/*
			if werr := f.writeq.wait(ctx, &f.writeqEntry, flowPriority); werr != nil {
				err = io.EOF
				break
			}

			fmt.Fprintf(os.Stderr, "flow: writeMsg: %p:%p:%3v: %p:%p: after flow control and writeq: %v %v/%v (alsoclose: %v)\n", f.conn, f, f.id, f.writeq, &f.writeqEntry, tokens, sent, totalSize, alsoClose)

			// Actually write to the wire.  This is also where encryption
			// happens, so this part can be slow.
			d := &message.Data{ID: f.id, Payload: tosend}
			if alsoClose && len(parts) == 0 {
				d.Flags |= message.CloseFlag
			}
			if f.noEncrypt {
				d.Flags |= message.DisableEncryptionFlag
			}

			wasOpened := f.setOpened()
			fmt.Fprintf(os.Stderr, "flow: writeMsg: %p:%p:%3v: was opened: %v\n", f.conn, f, f.id, wasOpened)
			if wasOpened {
				err = f.conn.mp.writeMsg(ctx, d)
			} else {
				err = f.conn.mp.writeMsg(ctx, &message.OpenFlow{
					ID:              f.id,
					InitialCounters: f.flowControl.shared.bytesBufferedPerFlow,
					BlessingsKey:    bkey,
					DischargeKey:    dkey,
					Flags:           d.Flags,
					Payload:         d.Payload,
				})
			}
			f.writeq.done(&f.writeqEntry)
			if err != nil {
				break
			}
		*/

	}

	fmt.Fprintf(os.Stderr, "flow: writeMsg: %p:%p:%3v: %p:%p: done flow control: %v/%v (alsoclose: %v): error: %v\n", f.conn, f, f.id, f.writeq, &f.writeqEntry, sent, totalSize, alsoClose, err)

	if debug {
		f.ctx.Infof("finishing write on %d(%p): %v", f.id, f, err)
	}

	if alsoClose || err != nil {
		f.close(ctx, false, err)
	}
	return sent, err
}

func (f *flw) sendFlowMessage(ctx *context.T, wasOpened, alsoClose, finalPart bool, bkey, dkey uint64, payload [][]byte) error {
	if werr := f.writeq.wait(ctx, &f.writeqEntry, flowPriority); werr != nil {
		return io.EOF
	}

	// Actually write to the wire.  This is also where encryption
	// happens, so this part can be slow.
	d := &message.Data{ID: f.id, Payload: payload}
	if alsoClose && finalPart {
		d.Flags |= message.CloseFlag
	}
	if f.noEncrypt {
		d.Flags |= message.DisableEncryptionFlag
	}
	var err error
	if wasOpened {
		err = f.conn.mp.writeMsg(ctx, d)
	} else {
		err = f.conn.mp.writeMsg(ctx, &message.OpenFlow{
			ID:              f.id,
			InitialCounters: f.flowControl.shared.bytesBufferedPerFlow,
			BlessingsKey:    bkey,
			DischargeKey:    dkey,
			Flags:           d.Flags,
			Payload:         d.Payload,
		})
	}
	f.writeq.done(&f.writeqEntry)
	return err
}

// WriteMsg is like Write, but allows writing more than one buffer at a time.
// The data in each buffer is written sequentially onto the flow.  Returns the
// number of bytes written.  WriteMsg must return a non-nil error if it writes
// less than the total number of bytes from all buffers.
// Write, WriteMsg, and WriteMsgAndClose should not be called concurrently
// with themselves or each other.
func (f *flw) WriteMsg(parts ...[]byte) (int, error) {
	return f.writeMsg(false, parts...)
}

// WriteMsgAndClose performs WriteMsg and then closes the flow.
// Write, WriteMsg, and WriteMsgAndClose should not be called concurrently
// with themselves or each other.
func (f *flw) WriteMsgAndClose(parts ...[]byte) (int, error) {
	return f.writeMsg(true, parts...)
}

// SetContext sets the context associated with the flow.  Typically this is
// used to set state that is only available after the flow is connected, such
// as a more restricted flow timeout, or the language of the request.
// Calling SetContext may invalidate values previously returned from Closed.
//
// The flow.Manager associated with ctx must be the same flow.Manager that the
// flow was dialed or accepted from, otherwise an error is returned.
// TODO(mattr): enforce this restriction.
//
// TODO(mattr): update v23/flow documentation.
// SetContext may not be called concurrently with other methods.
func (f *flw) SetDeadlineContext(ctx *context.T, deadline time.Time) *context.T {
	f.lock()
	defer f.unlock()

	if f.closed {
		// If the flow is already closed, don't allocate a new
		// context, we might end up leaking it.
		return ctx
	}

	if f.cancel != nil {
		f.cancel()
	}
	if !deadline.IsZero() {
		f.ctx, f.cancel = context.WithDeadline(ctx, deadline)
	} else {
		f.ctx, f.cancel = context.WithCancel(ctx)
	}
	return f.ctx
}

// LocalEndpoint returns the local vanadium endpoint.
func (f *flw) LocalEndpoint() naming.Endpoint {
	return f.conn.local
}

// RemoteEndpoint returns the remote vanadium endpoint.
func (f *flw) RemoteEndpoint() naming.Endpoint {
	return f.remote
}

// RemoteAddr returns the remote address of the peer.
func (f *flw) RemoteAddr() net.Addr {
	return f.conn.remoteAddr
}

// LocalBlessings returns the blessings presented by the local end of the flow
// during authentication.
func (f *flw) LocalBlessings() security.Blessings {
	return f.localBlessings
}

// RemoteBlessings returns the blessings presented by the remote end of the
// flow during authentication.
func (f *flw) RemoteBlessings() security.Blessings {
	return f.remoteBlessings
}

// LocalDischarges returns the discharges presented by the local end of the
// flow during authentication.
//
// Discharges are organized in a map keyed by the discharge-identifier.
func (f *flw) LocalDischarges() map[string]security.Discharge {
	return f.localDischarges
}

// RemoteDischarges returns the discharges presented by the remote end of the
// flow during authentication.
//
// Discharges are organized in a map keyed by the discharge-identifier.
func (f *flw) RemoteDischarges() map[string]security.Discharge {
	return f.remoteDischarges
}

// Conn returns the connection the flow is multiplexed on.
func (f *flw) Conn() flow.ManagedConn {
	return f.conn
}

// Closed returns a channel that remains open until the flow has been closed remotely
// or the context attached to the flow has been canceled.
//
// Note that after the returned channel is closed starting new writes will result
// in an error, but reads of previously queued data are still possible.  No
// new data will be queued.
// TODO(mattr): update v23/flow docs.
func (f *flw) Closed() <-chan struct{} {
	return f.currentContext().Done()
}

// ID returns the ID of this flow.
func (f *flw) ID() uint64 {
	return f.id
}

// setOpened ensures that on exit f.opened is true and if necessary
// (ie. f.opened was false on entry) it will call Done on the unopenedFlows
// waitgroup.
func (f *flw) setOpened() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.setOpenedLocked()
}

func (f *flw) setOpenedLocked() bool {
	if f.opened {
		return true
	}
	f.conn.unopenedFlows.Done()
	f.opened = true
	return false
}

func (f *flw) close(ctx *context.T, closedRemotely bool, err error) {

	log := f.ctx.V(2)

	f.lock()
	closed := f.closed
	f.closed = true
	cancel := f.cancel
	f.unlock()

	if !closed {
		// delete the flow as soon as possible to ensure that flow control
		// releases are handled appropriately for a closed/closing flow.
		// From this point on, all flow control updates are handled as per
		// a closed flow.
		f.conn.deleteFlow(f.id)

		f.q.close(ctx)
		if log {
			ctx.Infof("closing %d(%p): %v", f.id, f, err)
		}
		cancel()
		// After cancel has been called no new writes will begin for this
		// flow.  There may be a write in progress, but it must finish
		// before another writer gets to use the channel.  Therefore we
		// can simply use sendMessage to send the close flow message.

		wasopened := f.setOpened()
		if wasopened && !closedRemotely && (f.conn.state() != Closing) {
			// Note: If the conn is closing there is no point in trying to
			// send the flow close message as it will fail.  This is racy
			// with the connection closing, but there are no ill-effects
			// other than spamming the logs a little so it's OK.
			serr := f.conn.sendMessage(ctx, false, expressPriority, &message.Data{
				ID:    f.id,
				Flags: message.CloseFlag,
			})
			if serr != nil && log {
				ctx.Infof("could not send close flow message: %v: close error (if any): %v", serr, err)
			}
		}
		// update flow control state now that we're guaranteed to no longer
		// send any messages. Messages may still arrive if the flow is closed
		// locally but not remotely, but they will be handled appropriately
		// with regard to flow control.
		f.flowControl.handleFlowClose(closedRemotely, f.conn.state() < Closing)
	}
}

// Close marks the flow as closed. After Close is called, new data cannot be
// written on the flow. Reads of already queued data are still possible.
func (f *flw) Close() error {
	f.close(f.currentContext(), false, nil)
	return nil
}

func (f *flw) markUsed() {
	if !f.sideChannel && f.id >= reservedFlows {
		f.conn.markUsed()
	}
}

// popFront removes the first num bytes from in and appends them to out
// returning in, out, and the actual number of bytes appended.
func popFront(in, out [][]byte, num int) ([][]byte, [][]byte, int) {
	i, sofar := 0, 0
	for i < len(in) && sofar < num {
		i, sofar = i+1, sofar+len(in[i])
	}
	out = append(out, in[:i]...)
	if excess := sofar - num; excess > 0 {
		i, sofar = i-1, num
		keep := len(out[i]) - excess
		in[i], out[i] = in[i][keep:], out[i][:keep]
	}
	return in[i:], out, sofar
}
