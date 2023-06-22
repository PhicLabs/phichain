package txpool

import (
	"sync"

	"github.com/PhicLabs/phichain/common"
	"github.com/PhicLabs/phichain/core/state"
)

// noncer is a tiny virtual state database to manage the executable nonces of
// accounts in the pool, falling back to reading from a real state database if
// an account is unknown.
type noncer struct {
	fallback *state.StateDB
	nonces   map[common.Address]uint64
	lock     sync.Mutex
}

// newNoncer creates a new virtual state database to track the pool nonces.
func newNoncer(statedb *state.StateDB) *noncer {
	return &noncer{
		fallback: statedb.Copy(),
		nonces:   make(map[common.Address]uint64),
	}
}

// get returns the current nonce of an account, falling back to a real state
// database if the account is unknown.
func (txn *noncer) get(addr common.Address) uint64 {
	// We use mutex for get operation is the underlying
	// state will mutate db even for read access.
	txn.lock.Lock()
	defer txn.lock.Unlock()

	if _, ok := txn.nonces[addr]; !ok {
		if nonce := txn.fallback.GetNonce(addr); nonce != 0 {
			txn.nonces[addr] = nonce
		}
	}
	return txn.nonces[addr]
}

// set inserts a new virtual nonce into the virtual state database to be returned
// whenever the pool requests it instead of reaching into the real state database.
func (txn *noncer) set(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()

	txn.nonces[addr] = nonce
}

// setIfLower updates a new virtual nonce into the virtual state database if the
// new one is lower.
func (txn *noncer) setIfLower(addr common.Address, nonce uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()

	if _, ok := txn.nonces[addr]; !ok {
		if nonce := txn.fallback.GetNonce(addr); nonce != 0 {
			txn.nonces[addr] = nonce
		}
	}
	if txn.nonces[addr] <= nonce {
		return
	}
	txn.nonces[addr] = nonce
}

// setAll sets the nonces for all accounts to the given map.
func (txn *noncer) setAll(all map[common.Address]uint64) {
	txn.lock.Lock()
	defer txn.lock.Unlock()

	txn.nonces = all
}
