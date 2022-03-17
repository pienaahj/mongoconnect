package mongoconnect

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// declare variables
var (
	Database *mongo.Database
	Client *mongo.Client
	Collection *mongo.Collection 
)

// Notes:
// To insert a document into a collection, first retrieve a Database and then 
// Collection instance from the Client:
// collection := client.Database("testdb").Collection("numbers")
// The Collection instance can then be used to insert documents:
// ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
// defer cancel()
// res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
// id := res.InsertedID

// ConfigDB populates the database variables by connecting to a
// mongo database(dbName) and collection(colName)with a 
// connection string(conStr) format:
// connectionStringAdmin string = "mongodb://admin:myadminpassword@192.168.0.148:27017"
// connectionStringUser string = "mongodb://user2:user2password@192.168.0.148:27017/user2?authSource=testdb"

// "mongodb://admin:myadminpassword@192.168.0.148:27017/dbName"

// This needs to be in the main.go file as a client expires with the main function
// func ConfigDB(conStr string, dbName string, colName string) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	var err error
// 	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(conStr))
// 	defer func() {
// 		if err = Client.Disconnect(ctx); err != nil {
// 			panic(err)
// 		}
// 	}()

// 	Database = Client.Database(dbName)
// 	Collection = Database.Collection(colName)
// }



// CheckConnection checks server connectivity using the Ping method
// Calling Connect does not block for server discovery. 
func CheckConnection(client *mongo.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println("Could not connect to mongo client")
		return false
	}
	return true
}

//  CreateEntry adds a record(doc) to the database(dbs) into Collection(collection)
func CreateEntry(dbs *mongo.Database, collection string, doc bson.D) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// get the database name
	dbName := dbs.Name()
	// res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	res, err := dbs.Collection(collection).InsertOne(ctx, doc)
	id := res.InsertedID
	if err != nil {
		return nil, fmt.Errorf("could not create record into : %s with error: %q",dbName, err)
	}
	return id, nil
}

//  CreateEntries adds a records(docs) to the database(dbs) into Collection(collection)
func CreateEntries(dbs *mongo.Database, collection string, docs []interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// get the database name
	dbName := dbs.Name()
	// Set the order option to false to allow operations to happen even if one of them errors
	opts := options.InsertMany().SetOrdered(false)
	res, err := dbs.Collection(collection).InsertMany(ctx, docs, opts)
	ids := res.InsertedIDs
	if err != nil {
		return nil, fmt.Errorf("could not create record into : %s with error: %q",dbName, err)
	}
	return ids, nil
}

// SingleItem returns a single item from the database
// For methods that return a single item, a SingleResult, which works like a *sql.Row:
// filter := bson.D{{"name", "pi"}}
func SingleItem(collection *mongo.Collection, filter bson.D) (bson.D, error) {
	// reserve momory for result
	var result bson.D

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		fmt.Println("record does not exist")
		return nil, fmt.Errorf("could not find record : %q with error: %q", filter, err)
	} else if err != nil {
		return nil, fmt.Errorf("an error:%q occured while finding filter : %q", err, filter)
	}
	// Do something with result...

	return result, nil
}

// AllItems retrieves all items in a collection
func AllItems(dbs *mongo.Database, colName string) ([]bson.M, error) {
	// reserve momory for result
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := dbs.Collection(colName).Find(ctx, bson.D{{}})
	if err != nil { 
		return nil, fmt.Errorf("an error:%q occured while finding all items", err)
	}
	defer cur.Close(ctx)
	var results []bson.M

	// To decode into result, use cursor.All()
	err = cur.All(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("an error:%q occured while decoding all items", err)
	}

	// To get the raw bson bytes use cursor.Current
	// raw := cur.Current
	// do something with raw...
	
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("an error:%q occured on cursor", err)
	}
	return results, nil
}

// RemoveOne deletes a record from a collection
func RemoveOne(collection *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	// create an expiring context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// set options to eliminate case sensitivity ie."name" filed is "Bob" or "bob"
	opts := options.Delete().SetCollation(&options.Collation{
		Locale: "en_US",
		Strength: 1,
		CaseLevel: false,
	})
	res, err := collection.DeleteOne(ctx, filter, opts)
	if err != nil {
		return nil, errors.New("could not delete record from mongodb ")
	}
	return res, nil
}

// RemoveMany deletes multiple record from a collection
func RemoveMany(collection *mongo.Collection, filter interface{}) (*mongo.DeleteResult, error) {
	// create an expiring context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// set options to eliminate case sensitivity ie."name" filed is "Bob" or "bob"
	opts := options.Delete().SetCollation(&options.Collation{
		Locale: "en_US",
		Strength: 1,
		CaseLevel: false,
	})
	res, err := collection.DeleteMany(ctx, filter, opts)
	if err != nil {
		return nil, errors.New("could not delete record from mongodb ")
	}
	return res, nil
}
