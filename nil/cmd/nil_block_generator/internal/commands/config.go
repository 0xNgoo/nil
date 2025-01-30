package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	configFilePath = "config.json"
	nilDBPath      = "test.db"
	fileMode       = 0o666
)

type Contract struct {
	Path    string `json:"path"`
	Address string `json:"address"`
}

func NewContract(path string, adr string) *Contract {
	return &Contract{
		Path:    path,
		Address: adr,
	}
}

type Call struct {
	ContractName string   `json:"contractName"`
	Method       string   `json:"method"`
	AbiPath      string   `json:"abiPath"`
	Address      string   `json:"address"`
	Args         []string `json:"args"`
	Count        int      `json:"count"`
}

func NewCall(contractName string, method, abiPath, address string, args []string, count int) *Call {
	return &Call{
		ContractName: contractName,
		Method:       method,
		AbiPath:      abiPath,
		Address:      address,
		Args:         args,
		Count:        count,
	}
}

type Config struct {
	WalletAdr  string              `json:"walletAdr"`
	PrivateKey string              `json:"privatekey"`
	Contracts  map[string]Contract `json:"conracts"`
	Calls      []Call              `json:"calls"`
}

func NewConfig() *Config {
	return &Config{
		WalletAdr:  "",
		PrivateKey: "",
		Contracts:  make(map[string]Contract),
	}
}

func CleanFiles() error {
	err := os.Remove(configFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = os.RemoveAll(nilDBPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ReadConfigFromFile() (*Config, error) {
	b, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func WriteConfigToFile(cfg *Config) error {
	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(configFilePath, jsonCfg, fileMode)
	return err
}

func InitConfig(adr, hexKey string) error {
	cfg := NewConfig()
	cfg.WalletAdr = adr
	cfg.PrivateKey = hexKey
	return WriteConfigToFile(cfg)
}

func AddContract(name string, path string, adr string) error {
	cfg, err := ReadConfigFromFile()
	if err != nil {
		return err
	}
	cfg.Contracts[name] = *NewContract(path, adr)
	return WriteConfigToFile(cfg)
}

func AddCall(contractName string, method string, args []string, count int) error {
	cfg, err := ReadConfigFromFile()
	if err != nil {
		return err
	}
	contract, err := cfg.GetContract(contractName)
	if err != nil {
		return err
	}
	abiPath := contract.Path + ".abi"
	cfg.Calls = append(cfg.Calls, *NewCall(contractName, method, abiPath, contract.Address, args, count))
	return WriteConfigToFile(cfg)
}

func ShowContracts(writer io.StringWriter) error {
	cfg, err := ReadConfigFromFile()
	if err != nil {
		return err
	}
	for contractName, contract := range cfg.Contracts {
		_, err = writer.WriteString(fmt.Sprintf("%s: %s: %s\n", contractName, contract.Address, contract.Path))
		if err != nil {
			return err
		}
	}
	return nil
}

func ShowCalls(writer io.StringWriter) error {
	cfg, err := ReadConfigFromFile()
	if err != nil {
		return err
	}
	for _, call := range cfg.Calls {
		_, err = writer.WriteString(fmt.Sprintf("%s: %s, %d\n", call.ContractName, call.Method, call.Count))
		if err != nil {
			return err
		}
	}
	return nil
}

func (cfg *Config) GetContract(contractName string) (*Contract, error) {
	contract, ok := cfg.Contracts[contractName]
	if !ok {
		return nil, fmt.Errorf("Contract %s not deployed", contractName)
	}
	return &contract, nil
}
