package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sharidas/go-petshotel/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Roles = []string{"MANAGER", "STAFF", "CUSTOMER", "ADMIN"}

type UserModel struct {
	Username string
	Password string
	Email    string
	Role     string
	Token    string
}

type customError struct {
	Err error
}

func (cerror *customError) Error() string {
	return fmt.Sprintf("%v", cerror.Err)
}

// Connect to mongodb
func connectMongo() (*mongo.Client, error) {
	configData := &config.Configuration{}
	configData.ConfigParser()
	fmt.Println(configData)

	client, err := mongo.NewClient(options.Client().ApplyURI(configData.MongoURL))

	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}

	fmt.Printf("The type of db = %T\n", client)

	return client, nil
}

// This model helps to create user
func (usermodel *UserModel) CreateUser() (string, error) {

	if usermodel.Username == "" {
		cerror := &customError{Err: errors.New("Invalid Username")}
		return "Invalid Username", cerror.Err
	} else if usermodel.Password == "" {
		cerror := &customError{Err: errors.New("Invalid Password")}
		return "Invalid Password", cerror.Err
	} else if usermodel.Email == "" {
		cerror := &customError{Err: errors.New("Invalid Email")}
		return "Invalid Email", cerror.Err
	} else if usermodel.Role == "" {
		cerror := &customError{Err: errors.New("Invalid Role")}
		return "Invalid Role", cerror.Err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, connectErr := connectMongo()

	//fmt.Println(mongoClient, connectErr)

	if connectErr != nil {
		return "Couldn't create client for mongodb", connectErr
	}

	clientErr := mongoClient.Connect(ctx)

	if clientErr != nil {
		fmt.Println("Mongodb connect error: ", clientErr)
		return "Mongodb connection error", clientErr
	}

	defer mongoClient.Disconnect(ctx)

	// An assumption that pets-hotel is provided in the uri...
	mongoDB := mongoClient.Database("pet-hotel")
	petHotelCollection := mongoDB.Collection("user")

	insertResult, insertErr := petHotelCollection.InsertOne(ctx, bson.D{
		{"username", usermodel.Username},
		{"password", usermodel.Password},
		{"email", usermodel.Email},
		{"role", usermodel.Role},
		{"allow_login", ""},
		{"token", usermodel.Token},
	})

	if insertErr != nil {
		fmt.Println("Insert Error: ", insertErr)
	}

	fmt.Println("Inserted Data = ", insertResult)

	return "", nil
}

func (usermodel *UserModel) VerifyToken() (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, connectErr := connectMongo()

	if connectErr != nil {
		return "Couldn't create client for mongodb", connectErr
	}

	clientErr := mongoClient.Connect(ctx)

	if clientErr != nil {
		fmt.Println("Mongodb connect error: ", clientErr)
		return "Mongodb connection error", clientErr
	}

	defer mongoClient.Disconnect(ctx)

	// An assumption that pets-hotel is provided in the uri...
	mongoDB := mongoClient.Database("pet-hotel")
	petHotelCollection := mongoDB.Collection("user")

	projection := bson.D{
		{"email", 1},
		{"token", 1},
		{"_id", 0},
	}

	result := petHotelCollection.FindOne(ctx, bson.D{
		{"email", usermodel.Email},
		{"token", usermodel.Token},
	}, options.FindOne().SetProjection(projection))

	var fetchResult bson.M

	decodeErr := result.Decode(&fetchResult)
	if decodeErr != nil {
		fmt.Printf("Not able to decode the data\n")
		return "Error", decodeErr
	}

	if fetchResult["token"] == usermodel.Token {
		//Lets delete token from the document and
		//Update the user to login
		petHotelCollection.UpdateOne(
			ctx,
			bson.D{
				{"email", usermodel.Email},
				{"token", fetchResult["token"]},
			},
			bson.D{
				{"$unset", bson.M{"token": 1}},
				{"$set", bson.M{"allow_login": 1}},
			},
			options.Update().SetUpsert(true),
		)

	}
	return "Ok", nil
}

func (usermodel *UserModel) LoginUser() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, connectErr := connectMongo()

	if connectErr != nil {
		return connectErr
	}

	clientErr := mongoClient.Connect(ctx)

	if clientErr != nil {
		fmt.Println("Mongodb connect error: ", clientErr)
		return clientErr
	}

	defer mongoClient.Disconnect(ctx)

	// An assumption that pets-hotel is provided in the uri...
	mongoDB := mongoClient.Database("pet-hotel")
	petHotelCollection := mongoDB.Collection("user")

	projection := bson.D{
		{"email", 1},
		{"password", 1},
		{"_id", 0},
	}
	result := petHotelCollection.FindOne(ctx, bson.D{
		{"email", usermodel.Email},
		{"token", usermodel.Password},
	}, options.FindOne().SetProjection(projection))

	var fetchResult bson.M

	decodeErr := result.Decode(&fetchResult)

	if fetchResult["email"] == usermodel.Email && fetchResult["password"] == usermodel.Password {
		return nil
	}
	// return a custom error
}
