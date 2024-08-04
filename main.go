package main

// mongodb+srv://ayaaakinleye:2701Akin2000@cluster0.opv1wfb.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0
// main.go

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RFID represents the RFID tag information
type RFIDCard struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	TagID     string             `bson:"tag_id"`
	IsActive  bool               `bson:"is_active"`
	Amount    float32            `bson:"amount"`
	IsBinOpen bool               `bson:"is_open"`
	UpdatedAt time.Time          `bson:"updated_at"`
	CreatedAt time.Time          `bson:"created_at"`
}

func RFIDData(db *mongo.Client, collection string) *mongo.Collection {
	return db.Database("dustbin").Collection(collection)
}

func main() {
	r := gin.Default()

	// Connect to MongoDB
	db := OpenConnection(os.Getenv("MONGO_URI"))

	if db == nil {
		panic("Cannot connect to database")
	}

	defer func(ctx context.Context) {
		db.Disconnect(ctx)

	}(context.TODO())

	// Handler for RFID data
	r.POST("/rfid", func(c *gin.Context) {

		var rfid RFIDCard

		if err := c.BindJSON(&rfid); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if rfid.TagID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing RFID tag ID"})
			return
		}

		// Check RFID status
		var result RFIDCard
		err := RFIDData(db, "rfid").FindOne(c, bson.M{"tag_id": rfid.TagID}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {

				rfid.CreatedAt = time.Now()
				_, err := RFIDData(db, "rfid").InsertOne(c, rfid)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add new tag"})
					return
				}
				c.JSON(http.StatusNotFound, gin.H{"error": "tag cannot be added"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "tag not found"})
			return
		}

		if !rfid.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "tag is inactive"})
			return
		}

		rfid.Amount -= 35

		// Open waste bin
		_, err = RFIDData(db, "rfid").UpdateOne(
			c,
			bson.M{"tag_id": rfid.TagID},
			bson.M{"$set": bson.M{"is_open": true, "updated_at": time.Now(), "amount": rfid.Amount}},
			options.Update().SetUpsert(true),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot remove charge for waste"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Waste bin opened and charge processed"})
	})

	// Start server
	if err := r.Run(":5000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
