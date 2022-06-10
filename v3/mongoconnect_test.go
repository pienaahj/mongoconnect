package mongoconnect_test

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"testing"

	mc "github.com/pienaahj/mongoconnect"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// Test the connection to MongoDB
func TestCheckConnection(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		// create a mock client
		client := mt.Client

		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		res := mc.CheckConnection(client)
		assert.True(t, res)
	})
}

func TestSingleItem(t *testing.T) {
	//  create a test object
	object := bson.D{
		{"john", "testname1"},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		// create a user as a test BSON.D entry
		expectedObject := object

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"john", "testname1"},
		}))
		filter := object
		objectResponse, err := mc.SingleItem(mc.Collection, filter)
		assert.Nil(t, err)
		assert.Equal(t, expectedObject, objectResponse)
	})
}

func TestFindManyItems(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		// create the test objects
		testObj1 := bson.D{
			{"name", "john"},
			{"email", "testEmail1"},
		}
		testObj2 := bson.D{
			{"name", "john"},
			{"email", "testEmail2"},
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, testObj1)
		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, testObj2)
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		filter := bson.D{{"name", "john"}}
		objects, err := mc.FindManyItems(mc.Collection, filter)
		assert.Nil(t, err)
		assert.Equal(t, []bson.M{
			{"name": "john", "email": "testEmail1"},
			{"name": "john", "email": "testEmail2"},
		}, objects)
	})
}

func TestAllItems(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		// create the test objects
		testObj1 := bson.D{
			{"name", "john"},
			{"email", "testEmail1"},
		}
		testObj2 := bson.D{
			{"name", "john"},
			{"email", "testEmail2"},
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, testObj1)
		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, testObj2)
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		objects, err := mc.AllItems(mc.Collection)
		assert.Nil(t, err)
		assert.Equal(t, []bson.M{
			{"name": "john", "email": "testEmail1"},
			{"name": "john", "email": "testEmail2"},
		}, objects)
	})
}

func TestRemoveOne(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})
		filter := bson.D{{"name", "john"}}
		_, err := mc.RemoveOne(mc.Collection, filter)
		assert.Nil(t, err)

	})

	mt.Run("no document deleted", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 0}, {"acknowledged", true}, {"n", 0}})
		filter := bson.D{{"name", "john"}}
		_, err := mc.RemoveOne(mc.Collection, filter)
		assert.NotNil(t, err)

	})
}

func TestRemoveMany(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})

		filter := interface{}(bson.D{{"name", "john"}})
		_, err := mc.RemoveMany(mc.Collection, filter)
		fmt.Printf("%v\n", err)
		assert.Nil(t, err)

	})

	mt.Run("no document deleted", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 0}, {"acknowledged", true}, {"n", 0}})
		filter := interface{}(bson.D{{"name", "john"}})
		_, err := mc.RemoveMany(mc.Collection, filter)
		assert.NotNil(t, err)

	})
}
func TestCreateEntry(t *testing.T) {
	// create a new mock client
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	// create the document to be inserted
	id := primitive.NewObjectID()
	doc := bson.D{{"_id", id}, {"name", "john"}}

	// run a test on the mock client
	mt.Run("success", func(mt *mtest.T) {
		// specify the collection
		mc.Collection = mt.Coll
		// set up the mock
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		// pass the mocks into the function under test
		insertedID, err := mc.CreateEntry(mc.Collection, doc)
		// assert the expected outcomes
		assert.Nil(t, err)
		assert.Equal(t, id, insertedID)
	})

	mt.Run("custom error duplicate", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))
		// create an empty doc
		doc := bson.D{{}}

		insertedID, err := mc.CreateEntry(mc.Collection, doc)
		fmt.Printf("Returned error:%v\n", err)
		assert.Nil(t, insertedID)
		assert.NotNil(t, err)
		// is "duplicate key error"  in the returned error string?
		if !strings.Contains(fmt.Sprint(err), "duplicate key error") {
			t.Errorf("not duplicate key error: %v", err)
		}
	})

	mt.Run("simple error", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 0}})
		// create an empty doc
		doc := bson.D{{}}

		insertedID, err := mc.CreateEntry(mc.Collection, doc)
		assert.Nil(t, insertedID)
		assert.NotNil(t, err)
	})
}

func TestCreateEntries(t *testing.T) {
	// create a new mock client
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	// create the test objects to insert
	testObj1 := bson.D{
		{"name", "john"},
		{"email", "testEmail1"},
	}
	testObj2 := bson.D{
		{"name", "john"},
		{"email", "testEmail2"},
	}

	mt.Run("success", func(mt *mtest.T) {
		mc.Collection = mt.Coll
		objects := []interface{}{testObj1, testObj2}
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		results, err := mc.CreateEntries(mc.Collection, objects)
		assert.Nil(t, err)
		assert.NotNil(t, results)

	})
}
