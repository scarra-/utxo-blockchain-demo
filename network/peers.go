package network

import (
	"fmt"
	"slices"
	"sync"

	"go.uber.org/zap"
)

type PeersStorage struct {
	peersLock sync.RWMutex
	peers     []string
	Logger    *zap.SugaredLogger
	hostAddr  string
}

func NewPeersStorage(logger *zap.SugaredLogger, hostAddr string, knownPeers []string) *PeersStorage {
	peersList := []string{}
	for _, peerAddr := range knownPeers {
		if peerAddr != hostAddr {
			peersList = append(peersList, peerAddr)
		}
	}

	return &PeersStorage{
		peers:    peersList,
		hostAddr: hostAddr,
		Logger:   logger,
	}
}

func (ps *PeersStorage) Len() int {
	ps.peersLock.RLock()
	defer ps.peersLock.RUnlock()

	return len(ps.peers)
}

func (ps *PeersStorage) Add(peerAddr string) {
	ps.peersLock.Lock()
	defer ps.peersLock.Unlock()

	if peerAddr == ps.hostAddr {
		return
	}

	if slices.Contains(ps.peers, peerAddr) {
		return
	}

	ps.Logger.Infow("adding_new_peer",
		"addr", peerAddr,
	)

	ps.peers = append(ps.peers, peerAddr)
}

func (ps *PeersStorage) Delete(peerAddr string) {
	ps.peersLock.Lock()
	defer ps.peersLock.Unlock()

	for i, addr := range ps.peers {
		if peerAddr == addr {
			ps.peers = append(ps.peers[:i], ps.peers[i+1:]...)
			ps.Logger.Infow("deleted_peer",
				"addr", peerAddr,
			)

			break
		}
	}
}

func (ps *PeersStorage) ForEach(callback func(peer string)) {
	ps.peersLock.RLock()
	defer ps.peersLock.RUnlock()

	for _, val := range ps.peers {
		callback(val)
	}
}

func (ps *PeersStorage) First() (string, error) {
	if ps.Len() > 0 {
		return ps.peers[0], nil
	}

	return "", fmt.Errorf("peers storage is empty")
}

func (ps *PeersStorage) PeerExists(peerAddr string) bool {
	for _, addr := range ps.peers {
		if peerAddr == addr {
			return true
		}
	}

	return false
}
