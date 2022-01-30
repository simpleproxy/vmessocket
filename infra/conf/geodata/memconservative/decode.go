package memconservative

import (
	"errors"
	"io"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/vmessocket/vmessocket/common/platform/filesystem"
)

var (
	errFailedToReadBytes            = errors.New("failed to read bytes")
	errFailedToReadExpectedLenBytes = errors.New("failed to read expected length of bytes")
	errInvalidGeodataFile           = errors.New("invalid geodata file")
	errInvalidGeodataVarintLength   = errors.New("invalid geodata varint length")
	errCodeNotFound                 = errors.New("code not found")
)

func emitBytes(f io.ReadSeeker, code string) ([]byte, error) {
	count := 1
	isInner := false
	tempContainer := make([]byte, 0, 5)

	var result []byte
	var advancedN uint64 = 1
	var geoDataVarintLength, codeVarintLength, varintLenByteLen uint64 = 0, 0, 0

Loop:
	for {
		container := make([]byte, advancedN)
		bytesRead, err := f.Read(container)
		if err == io.EOF {
			return nil, errCodeNotFound
		}
		if err != nil {
			return nil, errFailedToReadBytes
		}
		if bytesRead != len(container) {
			return nil, errFailedToReadExpectedLenBytes
		}

		switch count {
		case 1, 3:
			if container[0] != 10 {
				return nil, errInvalidGeodataFile
			}
			advancedN = 1
			count++
		case 2, 4:
			tempContainer = append(tempContainer, container...)
			if container[0] > 127 {
				advancedN = 1
				goto Loop
			}
			lenVarint, n := protowire.ConsumeVarint(tempContainer)
			if n < 0 {
				return nil, errInvalidGeodataVarintLength
			}
			tempContainer = nil
			if !isInner {
				isInner = true
				geoDataVarintLength = lenVarint
				advancedN = 1
			} else {
				isInner = false
				codeVarintLength = lenVarint
				varintLenByteLen = uint64(n)
				advancedN = codeVarintLength
			}
			count++
		case 5:
			if strings.EqualFold(string(container), code) {
				count++
				offset := -(1 + int64(varintLenByteLen) + int64(codeVarintLength))
				f.Seek(offset, 1)
				advancedN = geoDataVarintLength
			} else {
				count = 1
				offset := int64(geoDataVarintLength) - int64(codeVarintLength) - int64(varintLenByteLen) - 1
				f.Seek(offset, 1)
				advancedN = 1
			}
		case 6:
			result = container
			break Loop
		}
	}
	return result, nil
}

func Decode(filename, code string) ([]byte, error) {
	f, err := filesystem.NewFileSeeker(filename)
	if err != nil {
		return nil, newError("failed to open file: ", filename).Base(err)
	}
	defer f.Close()

	geoBytes, err := emitBytes(f, code)
	if err != nil {
		return nil, err
	}
	return geoBytes, nil
}
