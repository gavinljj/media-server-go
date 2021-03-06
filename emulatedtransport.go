package mediaserver

import (
	"fmt"

	"github.com/chuckpreslar/emission"
	"github.com/notedit/media-server-go/sdp"
	native "github.com/notedit/media-server-go/wrapper"
)

type EmulatedTransport struct {
	transport native.PCAPTransportEmulator
	streams   map[string]*IncomingStream
	*emission.Emitter
}

func NewEmulatedTransport(pcap string) *EmulatedTransport {
	transport := &EmulatedTransport{}
	transport.transport = native.NewPCAPTransportEmulator()
	transport.streams = map[string]*IncomingStream{}
	transport.Emitter = emission.NewEmitter()
	transport.transport.Open(pcap)
	return transport
}

func (e *EmulatedTransport) SetRemoteProperties(audio *sdp.MediaInfo, video *sdp.MediaInfo) {
	properties := native.NewProperties()
	if audio != nil {
		num := 0
		for _, codec := range audio.GetCodecs() {
			item := fmt.Sprintf("audio.codecs.%d", num)
			properties.SetProperty(item+".codec", codec.GetCodec())
			properties.SetProperty(item+".pt", codec.GetType())
			if codec.HasRTX() {
				properties.SetProperty(item+".rtx", codec.GetRTX())
			}
			num = num + 1
		}
		properties.SetProperty("audio.codecs.length", num)

		num = 0
		for id, uri := range audio.GetExtensions() {
			item := fmt.Sprintf("audio.ext.%d", num)
			properties.SetProperty(item+".id", id)
			properties.SetProperty(item+".uri", uri)
			num = num + 1
		}
		properties.SetProperty("audio.ext.length", num)
	}

	if video != nil {
		num := 0
		for _, codec := range video.GetCodecs() {
			item := fmt.Sprintf("video.codecs.%d", num)
			properties.SetProperty(item+".codec", codec.GetCodec())
			properties.SetProperty(item+".pt", codec.GetType())
			if codec.HasRTX() {
				properties.SetProperty(item+".rtx", codec.GetRTX())
			}
			num = num + 1
		}
		properties.SetProperty("video.codecs.length", num)

		num = 0
		for id, uri := range video.GetExtensions() {
			item := fmt.Sprintf("video.ext.%d", num)
			properties.SetProperty(item+".id", id)
			properties.SetProperty(item+".uri", uri)
			num = num + 1
		}
		properties.SetProperty("video.ext.length", num)
	}

	e.transport.SetRemoteProperties(properties)

	native.DeleteProperties(properties)
}

func (e *EmulatedTransport) CreateIncomingStream(streamInfo *sdp.StreamInfo) *IncomingStream {

	incomingStream := NewIncomingStreamWithEmulatedTransport(e.transport, native.PCAPTransportEmulatorToReceiver(e.transport), streamInfo)

	e.streams[incomingStream.GetID()] = incomingStream

	incomingStream.Once("stopped", func() {
		delete(e.streams, incomingStream.GetID())
	})

	incomingStream.On("track", func(track *IncomingStreamTrack) {
		e.EmitSync("incomingtrack", track, incomingStream)
	})

	for _, track := range incomingStream.GetTracks() {
		e.EmitSync("incomingtrack", track, incomingStream)
	}

	return incomingStream
}

func (e *EmulatedTransport) Play(time uint64) bool {
	e.transport.Seek(time)
	return e.transport.Play()
}

func (e *EmulatedTransport) Resume() bool {
	return e.transport.Play()
}

func (e *EmulatedTransport) Pause() bool {
	return e.transport.Stop()
}

func (e *EmulatedTransport) Seek(time uint64) bool {
	e.transport.Seek(time)
	return e.transport.Play()
}

func (e *EmulatedTransport) Stop() {

	if e.transport == nil {
		return
	}

	for _, stream := range e.streams {
		stream.Stop()
	}

	e.streams = map[string]*IncomingStream{}

	e.transport.Stop()

	native.DeletePCAPTransportEmulator(e.transport)
	e.transport = nil

}
