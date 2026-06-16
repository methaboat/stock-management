package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"stock-management/backend/config"
)

// Context Keys
type ContextKey string

const UserContextKey ContextKey = "user"

// Global Mongo references
var (
	MongoClient *mongo.Client
	DB          *mongo.Database

	UserCol     *mongo.Collection
	ProductsCol *mongo.Collection
	StockCol    *mongo.Collection
	POCol       *mongo.Collection
	SOCol       *mongo.Collection
	RMACol      *mongo.Collection
	AuditCol    *mongo.Collection
	ReconCol    *mongo.Collection
)

// Global Structs
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Role         string             `bson:"role" json:"role"` // "Admin", "WarehouseStaff", "SalesRep"
}

type Location struct {
	Warehouse string `bson:"warehouse" json:"warehouse"`
	Aisle     string `bson:"aisle" json:"aisle"`
	Bin       string `bson:"bin" json:"bin"`
}

type AuditLog struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         string             `bson:"userId" json:"userId"`
	Username       string             `bson:"username" json:"username"`
	Action         string             `bson:"action" json:"action"`         // "Create", "Receive", "Reserve", "Ship", "Adjust", "Return"
	EntityType     string             `bson:"entityType" json:"entityType"` // "Product", "Stock", "PO", "SO", "RMA"
	EntityID       string             `bson:"entityId" json:"entityId"`
	SKU            string             `bson:"sku" json:"sku"`
	QuantityChange float64            `bson:"quantityChange" json:"quantityChange"`
	Note           string             `bson:"note" json:"note"`
	Timestamp      time.Time          `bson:"timestamp" json:"timestamp"`
}

type UserSession struct {
	Username string
	Role     string
}

// JSON Handlers error mapping
func JSONErrorBadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func JSONErrorUnauthorized(c *gin.Context, err error) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
}

func JSONErrorForbidden(c *gin.Context, err error) {
	c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
}

func JSONErrorDefault(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func JSONSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// Database Operations
func ConnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to MongoDB at %s...", config.C.MongoURI)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.C.MongoURI))
	if err != nil {
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Warning: Could not ping MongoDB. Proceeding, but database might be offline: %v", err)
	}

	MongoClient = client
	DB = client.Database(config.C.DatabaseName)

	UserCol = DB.Collection("users")
	ProductsCol = DB.Collection("products")
	StockCol = DB.Collection("stock_levels")
	POCol = DB.Collection("purchase_orders")
	SOCol = DB.Collection("sales_orders")
	RMACol   = DB.Collection("returns_rma")
	AuditCol = DB.Collection("audit_logs")
	ReconCol = DB.Collection("reconciliations")

	ConfigureIndexes()
	SeedUsers()

	return nil
}

func ConfigureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Unique indexes
	ProductsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sku", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	StockCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sku", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	POCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "poNumber", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	SOCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "soNumber", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	UserCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func SeedUsers() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := UserCol.CountDocuments(ctx, bson.M{})
	if err != nil || count > 0 {
		return
	}

	seeds := []struct {
		username string
		password string
		role     string
	}{
		{"admin", "adminpass", "Admin"},
		{"staff", "staffpass", "WarehouseStaff"},
		{"sales", "salespass", "SalesRep"},
	}

	for _, s := range seeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Warning: could not hash password for %s: %v", s.username, err)
			continue
		}
		UserCol.InsertOne(ctx, User{
			Username:     s.username,
			PasswordHash: string(hash),
			Role:         s.role,
		})
	}
	log.Println("Database seeded with default users (admin, staff, sales)")
}

// LogAudit logs inventory actions to the audit collection
func LogAudit(ctx context.Context, username, action, entityType, entityId, sku string, qtyChange float64, note string) {
	if username == "" {
		username = "system"
	}
	audit := AuditLog{
		UserID:         username,
		Username:       username,
		Action:         action,
		EntityType:     entityType,
		EntityID:       entityId,
		SKU:            sku,
		QuantityChange: qtyChange,
		Note:           note,
		Timestamp:      time.Now(),
	}
	AuditCol.InsertOne(ctx, audit)
}
