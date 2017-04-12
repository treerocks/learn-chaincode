/*
Copyright 33.cn Corp. 2017 All Rights Reserved.

Chaincode for Meidi Corp.

Chang history:


2017/4/11 init version.  


*/

package main


import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var transactionNo int = 0

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Bill struct {
	BillId		string
	Maker		string
	Acceptor	string
	Receiver	string //Receiver name	
	IssueDate	int64
	ExpireDate	int64
	RecBank		string //Name of receiver's Bank
	Amount		int //Amount of money 
	Type		int //0:taels 1:Business 
	Form		int //0:paper 1:electronic 
	Status		int //0: 1: 2: 3: 4: 5: 
}

// Record all operation include the create/change or the bill

type Transaction struct {

	BillId		string
	Operation	string
	BillStatus	int
	Time		int64
	ID		int
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Chaicode_Meidi Init")

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState("MeiDiFabric", []byte("This is the block chain for MeiDiBill system..."))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ChainCode_Meidi Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func (t *SimpleChaincode) createBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 11 {
		return shim.Error("createBill(): Incorrect number of arguments. Expecting 11")
	}

	var bill	Bill
	var t_issue	time.Time
	var t_expire	time.Time
	var err		error
	var amount	int
	var billtype	int
	var form	int
	var status	int

	layout := ""
	t_issue, err =  time.parse(layout, args[4])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to pharse the time, plese follow the fomat as" + layout + "\"}"
		return shim.Error(jsonResp)
	}

	t_expire, err =  time.parse(layout, args[5])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to pharse the time, plese follow the fomat " + layout + "\"}"
		return shim.Error(jsonResp)
	}

	amount, err = strconv.Atoi(args[7])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	billtype, err = strconv.Atoi(args[8])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	form, err = strconv.Atoi(args[9])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	status, err = strconv.Atoi(args[10])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	bill = {BillId: args[0], Maker: args[1], Acceptor: args[2], Receiver: args[3],IssueDate: t_issue.Unix(), ExpireDate: t_expire.Unix() , RecBank: args[6], Amount: amount, Type: billtype, Form: form, Status: status }

	err = writeBill(stub, bill)
}


// Change the Bill status
func (t *SimpleChaincode) changeBillStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}


// Write Bill info into the fabric
func writeBill(stub shim.ChaincodeStubInterface, bill Bill) error {

	billBytes, err := json.Marshal(&bill)
	if err != nil {
		return errors.New("writeBill():" + err.Error())
	}
	err = stub.PutState("bill"+bill.BillId, billBytes)
	if err != nil {
		return errors.New("writeBill(): PutState Error" + err.Error())
	}
	return nil
}

//Write the transaction info into the Fabric
func writeTransaction(stub shim.ChaincodeStubInterface, transaction Transaction) error {
	var tsId string
	tsBytes, err := json.Marshal(&transaction)
	if err != nil {
		return errors.New("writeTransaction():" + err.Error())
	}
	tsId = strconv.Itoa(transaction.ID)
	if err != nil {
		return errors.New("writeTransaction(): want Integer number")
	}
	err = stub.PutState("transaction"+tsId, tsBytes)
	if err != nil {
		return errors.New("writeTransaction(): PutState Error" + err.Error())
	}
	return nil
}

func getBillById(stub shim.ChaincodeStubInterface, id string) (Company, []byte, error) {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
