package protocol

import (
	"crypto/hmac"
	"crypto/md5"
	"hash"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/uuid"
)

const IDBytesLen = 16

type ID struct {
	uuid   uuid.UUID
	cmdKey [IDBytesLen]byte
}

type IDHash func(key []byte) hash.Hash

func DefaultIDHash(key []byte) hash.Hash {
	return hmac.New(md5.New, key)
}

func NewAlterIDs(primary *ID, alterIDCount uint16) []*ID {
	alterIDs := make([]*ID, alterIDCount)
	prevID := primary.UUID()
	for idx := range alterIDs {
		newid := nextID(&prevID)
		alterIDs[idx] = NewID(newid)
		prevID = newid
	}
	return alterIDs
}

func NewID(uuid uuid.UUID) *ID {
	id := &ID{uuid: uuid}
	md5hash := md5.New()
	common.Must2(md5hash.Write(uuid.Bytes()))
	common.Must2(md5hash.Write([]byte("c48619fe-8f02-49e0-b9e9-edf763e17e21")))
	md5hash.Sum(id.cmdKey[:0])
	return id
}

func nextID(u *uuid.UUID) uuid.UUID {
	md5hash := md5.New()
	common.Must2(md5hash.Write(u.Bytes()))
	common.Must2(md5hash.Write([]byte("16167dc8-16b6-4e6d-b8bb-65dd68113a81")))
	var newid uuid.UUID
	for {
		md5hash.Sum(newid[:0])
		if !newid.Equals(u) {
			return newid
		}
		common.Must2(md5hash.Write([]byte("533eff8a-4113-4b10-b5ce-0f5d76b98cd2")))
	}
}

func (id *ID) Bytes() []byte {
	return id.uuid.Bytes()
}

func (id ID) CmdKey() []byte {
	return id.cmdKey[:]
}

func (id *ID) Equals(another *ID) bool {
	return id.uuid.Equals(&(another.uuid))
}

func (id *ID) String() string {
	return id.uuid.String()
}

func (id *ID) UUID() uuid.UUID {
	return id.uuid
}
