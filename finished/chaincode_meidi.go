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

	// Write the state to the ledger
	err := stub.PutState("MeiDiFabric", []byte("This is the block chain for MeiDiBill system..."))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ChainCode_Meidi Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "createBill" {
		// Create the Bill 
		return t.createBill(stub, args)
	} else if function == "changeBillStatus" {
		// Change the status of Bill  
		return t.changeBillStatus(stub, args)
	} else if function == "queryBill" {
		//Query the Bill info from fabric 
		return t.queryBill(stub, args)
	} else if function == "queryTransaction" {
		//Query the Bill info from fabric 
		return t.queryTransaction(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
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
	var billBytes []byte

	layout := "2006-01-02 15:04:05"
	t_issue, err =  time.Parse(layout, args[4])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to pharse the time, plese follow the fomat as" + layout + "\"}"
		return shim.Error(jsonResp)
	}

	t_expire, err =  time.Parse(layout, args[5])
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

	bill = Bill{BillId:args[0], Maker: args[1], Acceptor: args[2], Receiver: args[3], IssueDate: t_issue.Unix(), ExpireDate: t_expire.Unix(), RecBank: args[6], Amount: amount, Type: billtype, Form: form, Status: status}

	err = writeBill(stub, bill)

	if err != nil {
		return shim.Error("createBill(): Fail to write the bill" + err.Error())
	}

	billBytes, err = json.Marshal(&bill)
	if err != nil {
		return shim.Error("createBill():" + err.Error())
	}

//Any operation of Bill will treat as transaction.

	transaction := Transaction{BillId: args[0], Operation: "Create", BillStatus: status, Time: time.Now().Unix(), ID: transactionNo }

	err = writeTransaction(stub, transaction)
	if err != nil {
		return shim.Error("createBill(): Fail to write the transaction" + err.Error())
	}

	transactionNo += 1

	return shim.Success(billBytes)

}


// Change the Bill status
func (t *SimpleChaincode) changeBillStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("changeBillStatus(): Incorrect number of arguments. Expecting 2")
	}

	var bill Bill

	status, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("changeBillStatus(): status want Integer number")
	}

	bill,_,err = getBillById(stub, args[0])
	if err != nil {

		return shim.Error("changeBillStatus(): Fail to get Bill" + err.Error())

	}
	bill.Status = status

	err = writeBill(stub, bill)

	if err != nil {
		shim.Error("changeBillStatus():fail to write Bill" + err.Error())
	}

	transaction := Transaction{BillId: args[0], Operation: "changeStatus", BillStatus: status, Time: time.Now().Unix(), ID: transactionNo }

	err = writeTransaction(stub, transaction)
	if err != nil {
		return shim.Error("changeBillStatus(): Fail to write the transaction" + err.Error())
	}

	transactionNo += 1

	return shim.Success(nil)
}

// Query Bill info 
func (t *SimpleChaincode) queryBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("queryBill(): Incorrect number of arguments. Expecting 1")
	}

	_,billBytes,err := getBillById(stub,args[0])

	if err != nil {
		shim.Error("queryBill(): Fail to get Bill" +  err.Error())
	}

	return shim.Success(billBytes)

}

// Query Transaction info 
func (t *SimpleChaincode) queryTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("queryTransaction(): Incorrect number of arguments. Expecting 1")
	}

	_,transBytes,err := getTransactionById(stub,args[0])

	if err != nil {
		return shim.Error("queryBill(): Fail to get Bill" +  err.Error())
	}

	return shim.Success(transBytes)

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

func getBillById(stub shim.ChaincodeStubInterface, id string) (Bill, []byte, error) {
	var bill Bill
	billBytes, err := stub.GetState("bill" + id)
	if err != nil {
		fmt.Println("Error retrieving cpBytes")
	}
	err = json.Unmarshal(billBytes, &bill)
	if err != nil {
		fmt.Println("Error unmarshalling Bill")
	}
	return bill, billBytes, nil
}



func getTransactionById(stub shim.ChaincodeStubInterface, id string) (Transaction, []byte, error) {
	var transaction Transaction
	transBytes, err := stub.GetState("transaction" + id)
	if err != nil {
		fmt.Println("Error retrieving transBytes")
	}
	err = json.Unmarshal(transBytes, &transaction)
	if err != nil {
		fmt.Println("Error unmarshalling transaction")
	}
	return transaction, transBytes, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
