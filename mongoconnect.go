package mongoconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// declare variables
var (
	database *mongo.Database
	client *mongo.Client
	collection *mongo.Collection

)



// ConfigDB populates the database variables by connecting to a
// mongo database(dbName) and collection(colName)
func ConfigDB(dbName string, colName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:myadminpassword@192.168.0.148:27017/" + dbName))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	database = client.Database(dbName)
	collection = database.Collection(colName)
}



// CheckConnection checks server connectivity using the Ping method
// Calling Connect does not block for server discovery. 
func CheckConnection(dbs *mongo.Database) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := dbs.Client().Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Printf("could not connect to db : %s", dbs.Name())
		return false
	}
	return true
}


// To insert a document into a collection, first retrieve a Database and then 
// Collection instance from the Client:

// collection := client.Database("testdb").Collection("numbers")

//  CreateEntry adds a record(doc) to the database(dbs) into Collection(collection)
func CreateEntry(dbs *mongo.Database, collection string, doc bson.D) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// gert the database name
	dbName := dbs.Name()
	// res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	res, err := dbs.Collection(collection).InsertOne(ctx, doc)
	id := res.InsertedID
	if err != nil {
		return nil, fmt.Errorf("could not create record into : %s with error: %q",dbName, err)
	}
	return id, nil
}


// The Collection instance can then be used to insert documents:
// ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
// defer cancel()
// res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
// id := res.InsertedID


// Several query methods return a cursor, which can be used like this:
// MoreItems returns more than one item from the database
func MoreItems(dbs *mongo.Database, collection string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cur, err := dbs.Collection(collection).Find(ctx, bson.D{})
	if err != nil { log.Fatal(err) }
	defer cur.Close(ctx)
	// make the results slice
	type result struct{
		Foo string
		Bar int32
	}
	var results []result
	for cur.Next(ctx) {
	// To decode into a struct, use cursor.Decode()
	result := struct{
		Foo string
		Bar int32
	}{}
	err := cur.Decode(&result)
	if err != nil { log.Fatal(err) }
	// do something with result...
	results = append(results, result)
	// To get the raw bson bytes use cursor.Current
	// raw := cur.Current
	// do something with raw...
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}


// SingleItem returns a single item from the database
// For methods that return a single item, a SingleResult, which works like a *sql.Row:
// filter := bson.D{{"name", "pi"}}
func SingleItem(filter bson.D) (interface{}, error) {
	// reserve momory for result
	var result struct {
    	Value float64
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		fmt.Println("record does not exist")
		return nil, fmt.Errorf("could not find record : %q with error: %q", filter, err)
	} else if err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("an error:%q occured while finding filter : %q", err, filter)
	}
	// Do something with result...

	return result, nil
}
// DeleteEntry removes a item from the database
func DeleteEntry(dbs *mongo.Database, colName string, filter bson.D) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// get the collection
	coll := dbs.Collection(colName)

	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Printf("an error occured while tryng to delete %q", filter)
		return false
	}
	fmt.Printf("record delete : %q", result)
	return true
}