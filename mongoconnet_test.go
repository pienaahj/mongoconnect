package mongoconnect_test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"testing"

	mc "github.com/pienaahj/mongoconnect"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

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
