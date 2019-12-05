
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"strconv"
	"time"
)

type transferRequest struct {
	EmailID string `json:"email"`
	TransferAmount string `json:"transferAmount"`
}
type check struct{
	EmailID string `json:"email"`
	EmailID2 string `json:"email2"`
}
type error struct{
	Message string `json:"message"`
}
type recurrin struct{
	Email string `json:"email"`
	Operation string `json:"operation"`
	Type string `json:"type"`
	Duration string `json:"duration"`
	Amount string `json:"Amount"`
}
type transferAmount struct{
	Sender string `json:"email"`
	Receiver string `json:"email2"`
	TransferAmount string `json:"transferAmount"`
}
type accounts struct {
	Type string `bson:"type"`
	Email string `bson:"email"`
	Balance string `bson:"balance"`
	Date string `bson:"date"`
}
func main(){
	router := mux.NewRouter()
	router.HandleFunc("/transfer",transferGet).Methods("GET")
	router.HandleFunc("/transfer",transferPut).Methods("PUT")
	router.HandleFunc("/recurring",recurringPost).Methods("POST")
	router.HandleFunc("/recurring",recurringGet).Methods("GET")
	router.HandleFunc("/transferWithinBank",transferWithinBank).Methods("PUT")
	router.HandleFunc("/admin",addtransactionsAdmin).Methods("PUT")
	log.Fatal(http.ListenAndServe(":3000", router))
}

func transferGet(w http.ResponseWriter,r *http.Request,) {
	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Agrifund?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	collection := client.Database("test").Collection("account")
	var result transferRequest
	locationId := mux.Vars(r)["EmailID"]
	err = collection.FindOne(context.TODO(), bson.D{{"EmailID", locationId}}).Decode(&result)
	if err != nil {
		fmt.Println("accounts document error")
		return
	}

	fmt.Printf("Found a document: %+v\n", result)

}

func addtransactionsAdmin(w http.ResponseWriter,r *http.Request,){
	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Bank?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	var req transferRequest
	var Check check
	err2:=json.NewDecoder(r.Body).Decode(&req)
	if(err2!=nil){
		log.Fatal(err2)
	}
	Check.EmailID=req.EmailID

	obj,err:=json.Marshal(map[string]string{
		"email":req.EmailID,
		"type":"savings",
		"operation":"credit",
		"amount":req.TransferAmount,
	})
	requestBody,err:=json.Marshal(map[string]json.RawMessage{
		"MessageBody":obj,
	})
	fmt.Println(requestBody)
	if(err!=nil){
		log.Fatal(err)
	}
	resp,err:=http.Post("https://4l0u135eh3.execute-api.us-east-1.amazonaws.com/test/api/send","application/json",bytes.NewBuffer(requestBody))
	if(err!=nil){
		log.Fatal(err)
	}
	//fmt.Println(resp.Body)
	fmt.Println(resp)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("Transaction successfull")
}


func transferWithinBank(w http.ResponseWriter,r *http.Request,){
	var result accounts
	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Bank?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	var req transferAmount
	var Check check
	err2 := json.NewDecoder(r.Body).Decode(&req)
	if(err2 != nil){
		log.Fatal(err2)
	}

	collection := client.Database("Bank").Collection("accounts")
	collection.FindOne(context.TODO(),bson.D{{"email",req.Sender}}).Decode(&result)
	fmt.Println(result)
	balance,err:=strconv.Atoi(result.Balance)
	requestedAmount,_:=strconv.Atoi(req.TransferAmount)
	if balance<requestedAmount {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode("Insufficient Balance!")
		return
	}
	Check.EmailID = req.Sender
	obj, err := json.Marshal(map[string]string{
		"email" : req.Sender,
		"type" : "savings",
		"operation" : "debit",
		"amount" : req.TransferAmount,
	})
	requestBody, err := json.Marshal(map[string]json.RawMessage{
		"MessageBody" : obj,
	})
	fmt.Println(requestBody)
	if(err != nil){
		log.Fatal(err)
	}
	resp,err := http.Post("https://4l0u135eh3.execute-api.us-east-1.amazonaws.com/test/api/send","application/json",bytes.NewBuffer(requestBody))
	if(err != nil){
		log.Fatal(err)
	}
	fmt.Println(resp.Body)

	Check.EmailID2 = req.Receiver

	obj, err = json.Marshal(map[string]string{
		"email" : req.Receiver,
		"type" : "savings",
		"operation" : "credit",
		"amount" : req.TransferAmount,
	})
	requestBody, err = json.Marshal(map[string]json.RawMessage{
		"MessageBody" : obj,
	})
	fmt.Println(requestBody)
	if(err != nil){
		log.Fatal(err)
	}
	resp,err = http.Post("https://4l0u135eh3.execute-api.us-east-1.amazonaws.com/test/api/send","application/json",bytes.NewBuffer(requestBody))
	if(err != nil){
		log.Fatal(err)
	}
	//fmt.Println(resp.Body)
	fmt.Println(resp)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("Transaction successfull")
}

func transferPut(w http.ResponseWriter,r *http.Request,){

	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Agrifund?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	var req transferRequest
	var Check check
	err2:=json.NewDecoder(r.Body).Decode(&req)
	if(err2!=nil){
		log.Fatal(err2)
	}
	Check.EmailID=req.EmailID

	obj,err:=json.Marshal(map[string]string{
		"email":req.EmailID,
		"type":"savings",
		"operation":"debit",
		"amount":req.TransferAmount,
	})
	requestBody,err:=json.Marshal(map[string]json.RawMessage{
		"MessageBody":obj,
	})
	fmt.Println(requestBody)
	if(err!=nil){
		log.Fatal(err)
	}
	resp,err:=http.Post("https://4l0u135eh3.execute-api.us-east-1.amazonaws.com/test/api/send","application/json",bytes.NewBuffer(requestBody))
	if(err!=nil){
		log.Fatal(err)
	}
	fmt.Println(resp.Body)
}


func recurringPost(w http.ResponseWriter,r *http.Request) {
	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Agrifund?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	var req recurrin
	err2 := json.NewDecoder(r.Body).Decode(&req)
	if (err2 != nil) {
		log.Fatal(err2)
	}

	collection := client.Database("test").Collection("recurringTransfer")
	_, err = collection.InsertOne(context.TODO(), req)
	if (err != nil) {
		log.Fatal(err)
	}
}

func recurringGet(w http.ResponseWriter, r* http.Request){

	var client *mongo.Client
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://nivali:Niv12345@agrifund-fqagq.mongodb.net/Agrifund?retryWrites=true&w=majority")
	fmt.Println("Client Options set...")
	client, err := mongo.Connect(ctx, clientOptions)
	fmt.Println("Mongo Connected...")
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		fmt.Println("error")
	}
	var req check
	err2 := json.NewDecoder(r.Body).Decode(&req)
	if (err2 != nil) {
		log.Fatal(err2)
	}

	var results []recurrin
	collection := client.Database("test").Collection("recurringTransfer")
	cursor,err:=collection.Find(context.TODO(),bson.D{{"email",req.EmailID}})
	if(err!=nil){
		log.Fatal(err)
	}
	var result recurrin
	for cursor.Next(context.TODO()){
		err=cursor.Decode(&result)
		if(err!=nil){
			log.Fatal(err)
		}
		results = append(results, result)
	}
	fmt.Println(results)
json.NewEncoder(w).Encode(results)
}