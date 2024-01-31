package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/aadejanovs/blockchain-demo/blockchain"
	"go.uber.org/zap"

	SYS "syscall"

	DEATH "github.com/vrecan/death/v3"
)

type ServerSettings struct {
	Protocol      string
	Version       int
	MsgNameLength int
	NodeID        string
	NodeAddress   string
	IsMiner       bool
	MinerAddress  string
	BlockTime     time.Duration
}

type Server struct {
	ServerSettings

	Logger       *zap.SugaredLogger
	chain        *blockchain.Blockchain
	client       *Client
	PeersStorage *PeersStorage
	Mempool      *Mempool
}

func NewServer(nodeID, minerAddress string) *Server {
	logger, err := blockchain.SetupLogger(nodeID)
	if err != nil {
		log.Fatal(err)
	}

	serverAddr := fmt.Sprintf("localhost:%s", nodeID)
	foundingNodeAddr := "localhost:3000"

	knownPeers := []string{}
	if serverAddr != foundingNodeAddr {
		knownPeers = append(knownPeers, foundingNodeAddr)
	}

	server := &Server{
		Logger:       logger,
		client:       NewClient(logger, serverAddr),
		chain:        blockchain.ContinueBlockchain(nodeID),
		PeersStorage: NewPeersStorage(logger, serverAddr, knownPeers),

		Mempool: NewMemPool(logger),
		ServerSettings: ServerSettings{
			NodeID:        nodeID,
			NodeAddress:   serverAddr,
			Protocol:      "tcp",
			Version:       1,
			MsgNameLength: 32,
			BlockTime:     time.Second * 5,
		},
	}

	server.Logger.Infow("known_peers",
		"peers", server.PeersStorage.peers,
	)

	go CloseDB(server.chain)

	if minerAddress != "" {
		server.IsMiner = true
		server.MinerAddress = minerAddress
	}

	return server
}

func (s *Server) Start() {
	ln, err := net.Listen(s.Protocol, s.NodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	firstPeer, err := s.PeersStorage.First()
	s.Logger.Infow("starting_node",
		"node_addr", s.NodeAddress,
		"first_peer", firstPeer,
		"error", err,
	)
	if err == nil {
		s.client.SendVersion(firstPeer, s.chain)
		s.client.SendGetMempoolTxs(firstPeer)
	}

	if s.IsMiner {
		go s.StartMining()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}

		go s.HandleConnection(conn)
	}
}

func (s *Server) StartMining() {
	ticker := time.NewTicker(s.BlockTime)

	s.Logger.Infow("mining_process_started",
		"block_time", s.BlockTime,
	)

	for {
		<-ticker.C
		s.MineTx()
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	req, err := io.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}

	msgName := BytesToMsg(req[:s.MsgNameLength])

	switch msgName {
	case msgVersion:
		s.HandleVersion(req)
	case msgAddresses:
		s.HandleAddresses(req)

	case msgBlockCreated:
		s.HandleBlockCreated(req)
	case msgGetBlock:
		s.HandleGetBlock(req)
	case msgGetBlockByHeight:
		s.HandleBlockByHeightRequested(req)
	case msgBlock:
		s.HandleBlock(req)

	case msgTx:
		s.HandleTx(req)

	case msgGetMempoolTxs:
		s.HandleGetMempoolTxs(req)
	default:
		s.Logger.Errorf("unkown_message_received",
			"name", msgName,
		)
	}
}

func CloseDB(chain *blockchain.Blockchain) {
	d := DEATH.NewDeath(SYS.SIGINT, SYS.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		chain.Logger.Infof("new_death_captured")

		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
