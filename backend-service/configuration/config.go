package configuration

import (
	"serart_be/data"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	Models *data.Models
}

var instance *Application
var once sync.Once
var db *mongo.Client

func New(pool *mongo.Client) *Application {
	db = pool
	return GetInstance()
}

func GetInstance() *Application {
	once.Do(func() {
		instance = &Application{
			Models: data.New(db),
		}
	})
	return instance
}
