package main

import (
	"math/rand"

	"go.mongodb.org/mongo-driver/mongo"
)

func createUniqueRandomString(n int32) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func saveToStorage(db, collection string, data interface{}) (*mongo.InsertOneResult, error) {
	c, err := App.mongo.InsertOne(db, collection, data)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func saveManyToStorage(db, collection string, data []interface{}) (*mongo.InsertManyResult, error) {
	c, err := App.mongo.InsertMany(db, collection, data)
	if err != nil {
		return nil, err
	}
	return c, nil
}
