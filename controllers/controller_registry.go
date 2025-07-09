package controllers

import "gorm.io/gorm"


var db *gorm.DB
func InitControllers(database *gorm.DB) {

	db = database
}
