package routes

import (
	"context"
	"fmt"
	"net/http"
	"server/server/models"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var validate = validator.New()
var orderCollection *mongo.Collection = OpenCollection(Client, "orders")

func AddOrder(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var order models.Order

	validationErr := validate.Struct(order)
	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validationErr.Error(),
		})
		return
	}
	order.ID = primitive.NewObjectID()
	result, insertErr := orderCollection.InsertOne(ctx, order)
	if insertErr != nil {
		msg := fmt.Sprintf("order item was not created")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": msg,
		})
		return
	}

	defer cancel()
	c.JSON(http.StatusOK, result)
}

func GetOrders(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var orders []bson.M

	cursor, err := orderCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer cancel()

	fmt.Println(orders)
	c.JSON(http.StatusOK, orders)
}

func GetOrdersByWaiter(c *gin.Context) {
	waiter := c.Params.ByName("waiter")
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var orders []bson.M

	cursor, err := orderCollection.Find(ctx, bson.M{
		"server": waiter,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err = cursor.All(ctx, &orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer cancel()

	fmt.Println(orders)
	c.JSON(http.StatusOK, orders)
}

func GetOrderById(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var order bson.M

	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = orderCollection.FindOne(ctx, bson.M{
		"_id": docID,
	}).Decode(&order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	defer cancel()
	fmt.Println(order)

	c.JSON(http.StatusOK, order)
}

func UpdateWaiter(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	type Waiter struct {
		Server *string `json:"server"`
	}
	var waiter Waiter

	result, err := orderCollection.UpdateOne(ctx, bson.M{"_id": docID},
		bson.D{
			{"$set", bson.D{{"server", waiter.Server}}},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	defer cancel()

	c.JSON(http.StatusOK, result.ModifiedCount)
}

func UpdateOrder(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var order models.Order

	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)

	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	validationErr := validate.Struct(order)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validationErr.Error(),
		})
		fmt.Println(validationErr)
		return
	}

	result, err := orderCollection.ReplaceOne(
		ctx,
		bson.M{"_id": docID},
		bson.M{
			"dish":   order.Dish,
			"price":  order.Price,
			"server": order.Server,
			"table":  order.Table,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	defer cancel()

	c.JSON(http.StatusOK, result.ModifiedCount)
}

func DeleteOrder(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	orderID := c.Params.ByName("id")
	docID, err := primitive.ObjectIDFromHex(orderID)

	result, err := orderCollection.DeleteOne(ctx, bson.M{
		"_id": docID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	defer cancel()

	c.JSON(http.StatusOK, result.DeletedCount)
}
