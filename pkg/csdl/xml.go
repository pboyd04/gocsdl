package csdl

import (
	"encoding/xml"
	"io"
	"sync"

	"github.com/orisano/gosax"
)

const gosaxBufferSize = 4 * 1024

type XMLDecoder interface {
	Decode(reader io.Reader, value any) error
}

type SaxXMLDecoder struct {
	saxReader *gosax.Reader
	decoder   *xml.Decoder
}

func NewSaxXMLDecoder() *SaxXMLDecoder {
	ret := &SaxXMLDecoder{
		saxReader: gosax.NewReaderSize(nil, gosaxBufferSize),
	}
	ret.decoder = xml.NewTokenDecoder(ret)
	return ret
}

func (d *SaxXMLDecoder) Decode(reader io.Reader, value any) error {
	d.saxReader.Reset(reader)
	d.saxReader.EmitSelfClosingTag = true
	err := d.decoder.Decode(value)
	d.saxReader.Reset(nil)
	if err != nil {
		// encoding/xml.Decoder has no reset API and may retain parsing state
		// after an error. Replace it before this SaxXMLDecoder is reused.
		d.decoder = xml.NewTokenDecoder(d)
	}
	return err
}

func (d *SaxXMLDecoder) Token() (xml.Token, error) {
	return gosax.TokenE(d.saxReader.Event())
}

var saxXMLDecoderPool = sync.Pool{
	New: func() any {
		return NewSaxXMLDecoder()
	},
}

type xmlDecodeFunc func(io.Reader, any) error

func decodeGosax(reader io.Reader, value any) error {
	decoder := saxXMLDecoderPool.Get().(*SaxXMLDecoder)
	err := decoder.Decode(reader, value)
	saxXMLDecoderPool.Put(decoder)
	return err
}

func decodeStandardXML(reader io.Reader, value any) error {
	return xml.NewDecoder(reader).Decode(value)
}
