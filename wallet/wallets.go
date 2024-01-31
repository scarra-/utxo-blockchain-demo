package wallet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const walletFile = "./tmp/wallets_%s.data"

type PersistedWallets struct {
	Wallets map[string][]byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func Load(nodeId string) (*Wallets, error) {
	wallets := Wallets{
		Wallets: make(map[string]*Wallet),
	}

	err := wallets.LoadFile(nodeId)

	return &wallets, err
}

func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	ws.Wallets[address] = wallet

	return address
}

func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) LoadFile(nodeId string) error {
	walletFile := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var persistedWallets Wallets

	err = json.Unmarshal(fileContent, &persistedWallets)

	if err != nil {
		return err
	}

	ws.Wallets = persistedWallets.Wallets

	return nil
}

func (ws *Wallets) SaveFile(nodeId string) {
	walletFile := fmt.Sprintf(walletFile, nodeId)

	b, err := json.Marshal(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, b, 0644)
	if err != nil {
		log.Panic(err)
	}
}
