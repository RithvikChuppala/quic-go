package quic

import (
	"errors"
	"sync"

	"github.com/quic-go/quic-go/internal/ackhandler"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/utils/ringbuffer"
	"github.com/quic-go/quic-go/internal/wire"
	"github.com/quic-go/quic-go/quicvarint"
)

type framer interface {
	HasData() bool

	QueueControlFrame(wire.Frame)
	AppendControlFrames([]ackhandler.Frame, protocol.ByteCount, protocol.VersionNumber) ([]ackhandler.Frame, protocol.ByteCount)

	AddActiveStream(protocol.StreamID)
	AppendStreamFrames([]ackhandler.StreamFrame, protocol.ByteCount, protocol.VersionNumber) ([]ackhandler.StreamFrame, protocol.ByteCount)

	Handle0RTTRejection() error
}

type framerI struct {
	mutex sync.Mutex

	streamGetter streamGetter

	activeStreams map[protocol.StreamID]struct{}
	streamQueue   ringbuffer.RingBuffer[protocol.StreamID]

	controlFrameMutex sync.Mutex
	controlFrames     []wire.Frame

	config         *Config
	streamMapPrio map[protocol.StreamID]int
	auxPriorSlice []int
}

var _ framer = &framerI{}

func newFramer(
		streamGetter streamGetter,
		config *Config) framer {
	return &framerI{
		streamGetter:  streamGetter,
		activeStreams: make(map[protocol.StreamID]struct{}),
		config: 	   config,
		streamMapPrio: make(map[protocol.StreamID]int),
	}
}

func (f *framerI) HasData() bool {
	f.mutex.Lock()
	hasData := !f.streamQueue.Empty()
	f.mutex.Unlock()
	if hasData {
		return true
	}
	f.controlFrameMutex.Lock()
	hasData = len(f.controlFrames) > 0
	f.controlFrameMutex.Unlock()
	return hasData
}

func (f *framerI) QueueControlFrame(frame wire.Frame) {
	f.controlFrameMutex.Lock()
	f.controlFrames = append(f.controlFrames, frame)
	f.controlFrameMutex.Unlock()
}

func (f *framerI) AppendControlFrames(frames []ackhandler.Frame, maxLen protocol.ByteCount, v protocol.VersionNumber) ([]ackhandler.Frame, protocol.ByteCount) {
	var length protocol.ByteCount
	f.controlFrameMutex.Lock()
	for len(f.controlFrames) > 0 {
		frame := f.controlFrames[len(f.controlFrames)-1]
		frameLen := frame.Length(v)
		if length+frameLen > maxLen {
			break
		}
		frames = append(frames, ackhandler.Frame{Frame: frame})
		length += frameLen
		f.controlFrames = f.controlFrames[:len(f.controlFrames)-1]
	}
	f.controlFrameMutex.Unlock()
	return frames, length
}

func (f *framerI) AddActiveStream(id protocol.StreamID) {
	f.mutex.Lock()
	if _, ok := f.activeStreams[id]; !ok {
		f.streamQueue.PushBack(id)
		f.activeStreams[id] = struct{}{}

		switch f.config.TypePrio {
			case "abs"://The stream queue is ordered by StreamPrior priorities slice.
				lenQ := f.streamQueue.Len()
				prior := 1
				//To assign priority to each slice in a map
				if v, ok := f.streamMapPrio[id]; ok {
					prior=v
				}else{
					fmt.Println("Else: \n",f.streamQueue, lenQ, prior,f.config.StreamPrio)
					if len(f.config.StreamPrio) > 0 {
						prior = f.config.StreamPrio[0]
						f.config.StreamPrio = f.config.StreamPrio[1:] //Delete the used priority for the next stream
					}
					f.streamMapPrio[id] = prior ///To assign priority to each slice in a map
				}
				f.auxPriorSlice = append(f.auxPriorSlice,prior)
				//Absolute priorization: the stream queue is ordered regarding the priorities of the stream (StreamPrio slice) which is also ordered
				posIni := lenQ-1
				newPrior := f.auxPriorSlice[posIni]
				var correctPos int //Correct position of the stream/prior regarding the prior
				for i := lenQ-1; i >= 0 ; i--{
					if  newPrior >= f.auxPriorSlice[i] {
						correctPos=i
					}
				}
				//To insert the stream ID and priority in the correct position
				auxSlice := append(f.auxPriorSlice[:correctPos], append([]int{newPrior}, f.auxPriorSlice[correctPos:posIni]...)...)
				copy(f.auxPriorSlice, auxSlice)
				f.streamQueue = append(f.streamQueue[:correctPos], append([]protocol.StreamID{id}, f.streamQueue[correctPos:posIni]...)...)
			case "rr": // stream ID has already been added
			default:
		}

	}
	f.mutex.Unlock()
}

func (f *framerI) AppendStreamFrames(frames []ackhandler.StreamFrame, maxLen protocol.ByteCount, v protocol.VersionNumber) ([]ackhandler.StreamFrame, protocol.ByteCount) {
	startLen := len(frames)
	var length protocol.ByteCount
	f.mutex.Lock()
	// pop STREAM frames, until less than MinStreamFrameSize bytes are left in the packet
	numActiveStreams := f.streamQueue.Len()
	for i := 0; i < numActiveStreams; i++ {
		if protocol.MinStreamFrameSize+length > maxLen {
			break
		}
		id := f.streamQueue.PopFront()
		// This should never return an error. Better check it anyway.
		// The stream will only be in the streamQueue, if it enqueued itself there.
		str, err := f.streamGetter.GetOrOpenSendStream(id)
		// The stream can be nil if it completed after it said it had data.
		if str == nil || err != nil {
			delete(f.activeStreams, id)
			continue
		}
		remainingLen := maxLen - length
		// For the last STREAM frame, we'll remove the DataLen field later.
		// Therefore, we can pretend to have more bytes available when popping
		// the STREAM frame (which will always have the DataLen set).
		remainingLen += quicvarint.Len(uint64(remainingLen))
		frame, ok, hasMoreData := str.popStreamFrame(remainingLen, v)
		if hasMoreData { // put the stream back in the queue (at the end)
			f.streamQueue.PushBack(id)
		} else { // no more data to send. Stream is not active
			delete(f.activeStreams, id)
		}
		// The frame can be "nil"
		// * if the receiveStream was canceled after it said it had data
		// * the remaining size doesn't allow us to add another STREAM frame
		if !ok {
			continue
		}
		frames = append(frames, frame)
		length += frame.Frame.Length(v)
	}
	f.mutex.Unlock()
	if len(frames) > startLen {
		l := frames[len(frames)-1].Frame.Length(v)
		// account for the smaller size of the last STREAM frame
		frames[len(frames)-1].Frame.DataLenPresent = false
		length += frames[len(frames)-1].Frame.Length(v) - l
	}
	return frames, length
}

func (f *framerI) Handle0RTTRejection() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.controlFrameMutex.Lock()
	f.streamQueue.Clear()
	for id := range f.activeStreams {
		delete(f.activeStreams, id)
	}
	var j int
	for i, frame := range f.controlFrames {
		switch frame.(type) {
		case *wire.MaxDataFrame, *wire.MaxStreamDataFrame, *wire.MaxStreamsFrame:
			return errors.New("didn't expect MAX_DATA / MAX_STREAM_DATA / MAX_STREAMS frame to be sent in 0-RTT")
		case *wire.DataBlockedFrame, *wire.StreamDataBlockedFrame, *wire.StreamsBlockedFrame:
			continue
		default:
			f.controlFrames[j] = f.controlFrames[i]
			j++
		}
	}
	f.controlFrames = f.controlFrames[:j]
	f.controlFrameMutex.Unlock()
	return nil
}
