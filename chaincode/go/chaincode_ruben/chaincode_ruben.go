/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright Receivership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type shard struct {
	ObjectType	string    `json:"docType"`	// docType is used to distinguish the various types of objects in state database
	Sender		string    `json:"Sender"`	// peer01.Org1
	ShardId		string    `json:"ShardId"`	// ShardId: Hash256{ IP, x, shard }
	DataId		string    `json:"DataId"`	// DataId: Hash256{ (c1,c1,···cN), K(r1,r1···rN), ShardId }
	Receiver	string    `json:"Receiver"`	// peer11.Org1, peer02.Org2, peer12.Org2

	Threshold	int 	  `json:"Threshold"`	// τ
	PUFNum		int 	  `json:"PUFNum"`	// 6 PUFs
	SuccessNum	int 	  `json:"SuccessNum"`
}


// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}


// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}


// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "addShard" {
		return t.addShard(stub, args)
	} else if function == "transferShard" {
	} else if function == "readShard" {
		return t.readShard(stub, args)
	} else if function == "queryShardsBySender" {
		return t.queryShardsBySender(stub, args)
	} else if function == "queryShards" {
		return t.queryShards(stub, args)
	} else if function == "getHistoryForShard" {
		return t.getHistoryForShard(stub, args)
	} else if function == "getShardsByRange" {
		return t.getShardsByRange(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)
	return shim.Error("Received unknown function invocation(Receive unknown function calls)")
}

// 1
// ============================================================
// addShard - create a new shard, store into chaincode state
// ============================================================
func (t *SimpleChaincode) addShard(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7(7parameters：Sender, ShardId, DataId, Receiver, Threshold, PUFNum, SuccessNum)")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init shard")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4st argument must be a non-empty string")
	}

	if len(args[4]) <= 0 {
		return shim.Error("5nd argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6rd argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7rd argument must be a non-empty string")
	}
	Sender := strings.ToLower(args[0])
	ShardId := args[1]
	DataId := strings.ToLower(args[2])
	Receiver := strings.ToLower(args[3])

	Threshold, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("5rd argument must be a numeric string")
	}
	PUFNum, err := strconv.Atoi(args[5])
	if err != nil {
		return shim.Error("6rd argument must be a numeric string")
	}
	SuccessNum, err := strconv.Atoi(args[6])
	if err != nil {
		return shim.Error("7rd argument must be a numeric string")
	}


	// ==== Check if shard already exists ====
	shardAsBytes, err := stub.GetState(ShardId)
	if err != nil {
		return shim.Error("Failed to get shard: " + err.Error())
	} else if shardAsBytes != nil {
		fmt.Println("This shard already exists: " + ShardId)
		return shim.Error("This shard already exists: " + ShardId)
	}

	// ==== Create shard object and marshal to JSON ====
	objectType := "shard"
	shard := &shard{objectType, Sender, ShardId, DataId, Receiver, Threshold, PUFNum, SuccessNum}
	shardJSONasBytes, err := json.Marshal(shard)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save shard to state ===
	err = stub.PutState(ShardId, shardJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the shard to enable Sender-based range queries, e.g. return all peer01.Org1 shards ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~Sender~ShardId.
	//  This will enable very efficient state range queries based on composite keys matching indexName~Sender~*
	indexName := "Sender~ShardId"
	senderShardIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{shard.Sender, shard.ShardId})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the shard.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(senderShardIdIndexKey, value)

	// ==== shard saved and indexed. Return success ====
	fmt.Println("- end init shard")
	return shim.Success(nil)
}


// 2
// ===========================================================
// transfer a shard by setting a new Receiver name on the shard
// ===========================================================
func (t *SimpleChaincode) transferShard(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	shardId := args[0]
	newReceiver := strings.ToLower(args[1])
	fmt.Println("- start transferShard ", shardId, newReceiver)

	shardAsBytes, err := stub.GetState(shardId)
	if err != nil {
		return shim.Error("Failed to get shard:" + err.Error())
	} else if shardAsBytes == nil {
		return shim.Error("shard does not exist")
	}

	shardToTransfer := shard{}
	err = json.Unmarshal(shardAsBytes, &shardToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	shardToTransfer.Receiver = newReceiver //change the Receiver

	shardJSONasBytes, _ := json.Marshal(shardToTransfer)
	err = stub.PutState(shardId, shardJSONasBytes) //rewrite the shard
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferShard (success)")
	return shim.Success(nil)
}


/*
// 3
// ==================================================
// delete - remove a shard key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var shardJSON shard
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	shardId := args[0]

	// to maintain the Sender~ShardId index, we need to read the shard first and get its Sender
	valAsbytes, err := stub.GetState(shardId) //get the shard from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + shardId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"shard does not exist: " + shardId + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &shardJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + shardId + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(shardId) //remove the shard from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "Sender~ShardId"
	senderShardIdIndexKey, err := stub.CreateCompositeKey(indexName, []string{shardJSON.Sender, shardJSON.ShardId})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(senderShardIdIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}
*/



// 4
// ===============================================
// readShard - read a shard from chaincode state
// ===============================================
func (t *SimpleChaincode) readShard(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var shardid, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting shardid of the shard to query")
	}

	shardid = args[0]
	valAsbytes, err := stub.GetState(shardid) //get the shardid from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + shardid + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"shard does not exist: " + shardid + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}







// =======Rich queries =========================================================================
// Two examples of rich queries are provided below: parameterized query & ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================
//
// ===== Example: Parameterized rich query =================================================
// queryShardsBySender queries for shards based on a passed in Receiver.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (Receiver).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
// 5
func (t *SimpleChaincode) queryShardsBySender(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "peer01.Org1"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	sender := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"shard\",\"Sender\":\"%s\"}}", sender)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}



// ===== Example: Ad hoc rich query ========================================================
// queryShards uses a query string to perform a query for shards.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryShardsForReceiver example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
// 6
func (t *SimpleChaincode) queryShards(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}
// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",\n")
		}
		buffer.WriteString("\n{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\n \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("\n}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]\n")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}



// 7
func (t *SimpleChaincode) getHistoryForShard(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	shardId := args[0]

	fmt.Printf("- start getHistoryForShard: %s\n", shardId)

	resultsIterator, err := stub.GetHistoryForKey(shardId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the shard
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",\n")
		}
		buffer.WriteString("\n{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(",\n \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON shard)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(",\n \"Timestamp\":")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString("\n}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]\n")
	fmt.Printf("- getHistoryForShard returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}



// ===========================================================================================
// getShardsByRange performs a range query based on the start and end keys provided.
// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
//
// *** Therefore, range queries are a safe option for performing update transactions based on query results. ***
// ===========================================================================================
// 8
func (t *SimpleChaincode) getShardsByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",\n")
		}
		buffer.WriteString("\n{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\n \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("\n}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]\n")

	fmt.Printf("- getShardsByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

