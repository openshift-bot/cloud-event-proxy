package hwevent

import (
	"io"

	"encoding/base64"
	"strconv"
	"sync"

	jsoniter "github.com/json-iterator/go"

	"github.com/redhat-cne/sdk-go/pkg/types"
)

var iterPool = sync.Pool{
	New: func() interface{} {
		return jsoniter.Parse(jsoniter.ConfigFastest, nil, 1024)
	},
}

func borrowIterator(reader io.Reader) *jsoniter.Iterator {
	iter := iterPool.Get().(*jsoniter.Iterator)
	iter.Reset(reader)
	return iter
}

func returnIterator(iter *jsoniter.Iterator) {
	iter.Error = nil
	iter.Attachment = nil
	iterPool.Put(iter)
}

// ReadJSON ...
func ReadJSON(out *Event, reader io.Reader) error {
	iterator := borrowIterator(reader)
	defer returnIterator(iterator)
	return readJSONFromIterator(out, iterator)
}

// ReadDataJSON ...
func ReadDataJSON(out *Data, reader io.Reader) error {
	iterator := borrowIterator(reader)
	defer returnIterator(iterator)
	return readDataJSONFromIterator(out, iterator)
}

// readJSONFromIterator allows you to read the bytes reader as an event
func readDataJSONFromIterator(out *Data, iterator *jsoniter.Iterator) error {
	var (
		// Universally parseable fields.
		version string
		data    []byte
		// These fields require knowledge about the specversion to be parsed.
		//schemaurl jsoniter.Any
	)

	for key := iterator.ReadObject(); key != ""; key = iterator.ReadObject() {
		// Check if we have some error in our error cache
		if iterator.Error != nil {
			return iterator.Error
		}

		// If no specversion ...
		switch key {
		case "version":
			version = iterator.ReadString()
		case "data":
			data = iterator.SkipAndReturnBytes()
		default:
			iterator.Skip()
		}
	}

	if iterator.Error != nil {
		return iterator.Error
	}
	out.Version = version
	out.Data = data
	return nil
}

// readJSONFromIterator allows you to read the bytes reader as an event
func readJSONFromIterator(out *Event, iterator *jsoniter.Iterator) error {
	var (
		// Universally parseable fields.
		id   string
		typ  string
		time *types.Timestamp
		data *Data
	)

	for key := iterator.ReadObject(); key != ""; key = iterator.ReadObject() {
		// Check if we have some error in our error cache
		if iterator.Error != nil {
			return iterator.Error
		}

		// If no specversion ...
		switch key {
		case "id":
			id = iterator.ReadString()
		case "type":
			typ = iterator.ReadString()
		case "time":
			time = readTimestamp(iterator)
		case "data":
			data, _ = readData(iterator)
		case "version":
		default:
			iterator.Skip()
		}
	}

	if iterator.Error != nil {
		return iterator.Error
	}
	out.Time = time
	out.ID = id
	out.Type = typ
	if data != nil {
		out.SetData(*data)
	}
	return nil
}

func readTimestamp(iter *jsoniter.Iterator) *types.Timestamp {
	t, err := types.ParseTimestamp(iter.ReadString())
	if err != nil {
		iter.Error = err
	}
	return t
}

func readData(iter *jsoniter.Iterator) (*Data, error) {
	data := &Data{
		Version: "",
		Data:    nil,
	}

	for key := iter.ReadObject(); key != ""; key = iter.ReadObject() {
		// Check if we have some error in our error cache
		if iter.Error != nil {
			return data, iter.Error
		}
		switch key {
		case "version":
			data.Version = iter.ReadString()
		case "data":
			data.Data = iter.SkipAndReturnBytes()
			unQuoted, err := strconv.Unquote(string(data.Data))
			if err != nil {
				return data, err
			}
			// []byte is encoded as a base64-encoded string with json.Marshal
			decoded, err := base64.StdEncoding.DecodeString(unQuoted)
			if err != nil {
				return data, err
			}
			data.Data = decoded
		default:
			iter.Skip()
		}
	}

	return data, nil
}

// UnmarshalJSON implements the json unmarshal method used when this type is
// unmarshaled using json.Unmarshal.
func (e *Event) UnmarshalJSON(b []byte) error {
	iterator := jsoniter.ConfigFastest.BorrowIterator(b)
	defer jsoniter.ConfigFastest.ReturnIterator(iterator)
	return readJSONFromIterator(e, iterator)
}

// UnmarshalJSON implements the json unmarshal method used when this type is
// unmarshaled using json.Unmarshal.
func (d *Data) UnmarshalJSON(b []byte) error {
	iterator := jsoniter.ConfigFastest.BorrowIterator(b)
	defer jsoniter.ConfigFastest.ReturnIterator(iterator)
	return readDataJSONFromIterator(d, iterator)
}
