package main

import "gitlab.faza.io/go-framework/logger"

func MongoMigrations() {
	_, err := App.mongo.AddUniqueIndex("test", "test-collection", "title")
	if err != nil {
		logger.Err(err.Error())
	}

	_, err = App.mongo.AddTextV3Index("test", "test-collection", "title")
	if err != nil {
		logger.Err(err.Error())
	}
}
