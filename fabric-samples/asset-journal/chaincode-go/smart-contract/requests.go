package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) RequestAccess(ctx contractapi.TransactionContextInterface, journalID string, access string) error {
	creator, err := GetClientID(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get owner ID, %v", err)
	}
	owner, _ := s.IsOwner(ctx, journalID)
	if owner {
		return fmt.Errorf("Cannot request additional access to a asset you own")
	}
	requests, err := AccessRequests(ctx, journalID)
	if err != nil {
		return fmt.Errorf("Failed to get Access requests %v", err)
	}
	requests.Entries[creator] = access
	requestsJSON, err := json.Marshal(requests)
	if err != nil {
		return fmt.Errorf("Failed json Marshal, %v", err)
	}
	return ctx.GetStub().PutState("request"+journalID, requestsJSON)
}

func AccessRequests(ctx contractapi.TransactionContextInterface, journalID string) (*Asset, error) {

	requestsJSON, err := ctx.GetStub().GetState("request" + journalID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	var requests Asset
	if requestsJSON != nil {
		err = json.Unmarshal(requestsJSON, &requests)
		if err != nil {
			return nil, err
		}
	} else {
		requests = Asset{Entries: make(map[string]string)}
	}
	return &requests, nil
}

func (s *SmartContract) GetAccessRequests(ctx contractapi.TransactionContextInterface, journalID string) (*Asset, error) {
	owner, _ := s.IsOwner(ctx, journalID)
	if !owner {
		return nil, fmt.Errorf("Only the owner can see requests")
	}

	requests, err := AccessRequests(ctx, journalID)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *SmartContract) AnswerAccessRequest(ctx contractapi.TransactionContextInterface, journalID string, peerID string, answer string) error {
	owner, err := s.IsOwner(ctx, journalID)
	if !owner {
		return fmt.Errorf("Only the owner can answer requests")
	}

	requests, err := AccessRequests(ctx, journalID)
	if err != nil {
		return err
	}
	if requests == nil {
		return fmt.Errorf("There are no access requests")
	}

	request := requests.Entries[peerID]
	switch answer {
	case "approve":
		{
			err = AddAuthentication(ctx, journalID, peerID, request)
			if err != nil {
				return err
			}
			delete(requests.Entries, peerID)
			requestJSON, err := json.Marshal(requests)
			if err != nil {
				return err
			}
			return ctx.GetStub().PutState("request"+journalID, requestJSON)
		}
	case "decline":
		{
			delete(requests.Entries, peerID)
			requestJSON, err := json.Marshal(requests)
			if err != nil {
				return err
			}
			return ctx.GetStub().PutState("request"+journalID, requestJSON)
			// delete request
		}
	default:
		return fmt.Errorf("The request access token is unrecognizable")
	}
	return nil
}
