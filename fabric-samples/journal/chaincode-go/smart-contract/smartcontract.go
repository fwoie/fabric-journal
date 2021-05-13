package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Patient
type SmartContract struct {
	contractapi.Contract
}

const (
	dateLayout = "01-02-2006"
)

// Asset is used both as a journal and list of users.
type Asset struct {
	Owner   string            `json:"owner"`
	Data    string            `json:"data"`
	Entries map[string]string `json:"entries"`
}

// CreateJournal creates the initial journal object
func (s *SmartContract) CreateJournal(ctx contractapi.TransactionContextInterface, journalID string) error {
	ok, err := JournalExists(ctx, journalID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("There is already a journal for that personal number")
	}

	creator, err := GetClientID(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get owner ID, %v", err)
	}

	// add creator to autherized users
	peers := Asset{
		Entries: map[string]string{creator: "r"},
	}
	peerJSON, err := json.Marshal(peers)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState("auth"+journalID, peerJSON)
	if err != nil {
		return fmt.Errorf("Failed to create autherization %v", err)
	}
	// create the journal with a default entry
	journal := Asset{
		Owner:   creator,
		Entries: make(map[string]string),
	}
	journal.Entries["default"] = "default"

	journalJSON, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(journalID, journalJSON)
}

// AddJournal adds a entry to the patient journal
func (s *SmartContract) AddEntry(ctx contractapi.TransactionContextInterface, journalID string, entryID string, data string) error {
	ok, err := Authenticate(ctx, journalID, "w")
	if !ok {
		return err
	}

	journal, err := GetJournal(ctx, journalID)
	if err != nil {
		return err
	}

	journal.Entries[entryID] = data

	journalJSON, err := json.Marshal(journal)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(journalID, journalJSON)
}

// GetJournal returns journals found in world state for the given personal number
func GetJournal(ctx contractapi.TransactionContextInterface, journalID string) (*Asset, error) {

	journalJSON, err := ctx.GetStub().GetState(journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if journalJSON == nil {
		return nil, fmt.Errorf("There are no journals searching for journalID %s", journalID)
	}

	var journal Asset
	err = json.Unmarshal(journalJSON, &journal)
	if err != nil {
		return nil, err
	}

	return &journal, nil
}

// ReadJournal returns journals found in world state for the given personal number
func (s *SmartContract) ReadJournal(ctx contractapi.TransactionContextInterface, journalID string) (*Asset, error) {
	ok, err := Authenticate(ctx, journalID, "r")
	if !ok {
		return nil, err
	}

	journalJSON, err := ctx.GetStub().GetState(journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if journalJSON == nil {
		return nil, fmt.Errorf("There are no journals searching for journalID %s", journalID)
	}

	var journal Asset
	err = json.Unmarshal(journalJSON, &journal)
	if err != nil {
		return nil, err
	}

	return &journal, nil
}

// GetEntry returns a single journal entry found in world state for the given personal number and ID
func (s *SmartContract) GetEntry(ctx contractapi.TransactionContextInterface, journalID string, entryID string) (*Asset, error) {
	journal, err := s.ReadJournal(ctx, journalID)
	if err != nil {
		return nil, err
	}

	data, ok := journal.Entries[entryID]
	if !ok {
		return nil, fmt.Errorf("There are no entries with the ID %v", entryID)
	}
	var entry Asset
	entry.Entries = make(map[string]string)
	entry.Entries[entryID] = data

	return &entry, nil
}

func JournalExists(ctx contractapi.TransactionContextInterface, journalID string) (bool, error) {

	journalJSON, err := ctx.GetStub().GetState(journalID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if journalJSON == nil {
		return false, nil
	}

	return true, nil
}

func GetClientID(ctx contractapi.TransactionContextInterface) (string, error) {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to get identity %v", err)
	}
	return clientID, nil
}

func AddAuthentication(ctx contractapi.TransactionContextInterface, journalID string, clientID string, access string) error {
	peers, err := GetAuthenticatedPeers(ctx, journalID)
	if err != nil {
		return err
	}

	peers[clientID] = access

	peersAsset := Asset{
		Entries: peers,
	}

	peerJSON, err := json.Marshal(peersAsset)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState("auth"+journalID, peerJSON)
}

func (s *SmartContract) IsOwner(ctx contractapi.TransactionContextInterface, journalID string) (bool, error) {
	clientID, err := GetClientID(ctx)
	if err != nil {
		return false, err
	}
	journal, err := GetJournal(ctx, journalID)
	if err != nil {
		return false, err
	}

	if journal.Owner == clientID {
		return true, nil
	}

	return false, fmt.Errorf("Peer is not the owner of this journal")
}

func Authenticate(ctx contractapi.TransactionContextInterface, journalID string, action string) (bool, error) {
	clientID, err := GetClientID(ctx)
	if err != nil {
		return false, err
	}
	peers, err := GetAuthenticatedPeers(ctx, journalID)
	if err != nil {
		return false, err
	}
	access, ok := peers[clientID]
	if ok {
		if access == "rw" || access == action {
			return true, nil
		}
	}
	return false, fmt.Errorf("Peer is not autherized to %s this journal", action)
}

func GetAuthenticatedPeers(ctx contractapi.TransactionContextInterface, journalID string) (map[string]string, error) {

	peersJSON, err := ctx.GetStub().GetState("auth" + journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if peersJSON == nil {
		return nil, fmt.Errorf("Failed to find autherized peers for personal number: ", journalID)
	}
	var peers Asset
	err = json.Unmarshal(peersJSON, &peers)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal Asset, ", err)
	}
	return peers.Entries, nil
}
