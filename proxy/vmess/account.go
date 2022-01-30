//go:build !confonly
// +build !confonly

package vmess

import (
	"strings"

	"github.com/vmessocket/vmessocket/common/dice"
	"github.com/vmessocket/vmessocket/common/protocol"
	"github.com/vmessocket/vmessocket/common/uuid"
)

type MemoryAccount struct {
	ID       *protocol.ID
	AlterIDs []*protocol.ID
	Security protocol.SecurityType

	AuthenticatedLengthExperiment bool
	NoTerminationSignal           bool
}

func (a *MemoryAccount) AnyValidID() *protocol.ID {
	if len(a.AlterIDs) == 0 {
		return a.ID
	}
	return a.AlterIDs[dice.Roll(len(a.AlterIDs))]
}

func (a *MemoryAccount) Equals(account protocol.Account) bool {
	vmessAccount, ok := account.(*MemoryAccount)
	if !ok {
		return false
	}
	return a.ID.Equals(vmessAccount.ID)
}

func (a *Account) AsAccount() (protocol.Account, error) {
	id, err := uuid.ParseString(a.Id)
	if err != nil {
		return nil, newError("failed to parse ID").Base(err).AtError()
	}
	protoID := protocol.NewID(id)
	var AuthenticatedLength, NoTerminationSignal bool
	if strings.Contains(a.TestsEnabled, "AuthenticatedLength") {
		AuthenticatedLength = true
	}
	if strings.Contains(a.TestsEnabled, "NoTerminationSignal") {
		NoTerminationSignal = true
	}
	return &MemoryAccount{
		ID:                            protoID,
		AlterIDs:                      protocol.NewAlterIDs(protoID, uint16(a.AlterId)),
		Security:                      a.SecuritySettings.GetSecurityType(),
		AuthenticatedLengthExperiment: AuthenticatedLength,
		NoTerminationSignal:           NoTerminationSignal,
	}, nil
}
