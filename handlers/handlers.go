package handlers

import (
	"fmt"
	db "people2/database"
	"people2/logging"
	"people2/models"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

var (
	log = logging.Config
)

// This API handler processes, checks, enriches and saves correct
// incoming messages to the database. Return a JSON success
// message or an error with its cause.
func Create(c *gin.Context) {
	f := logging.F()
	var dataMsg models.FullName
	if err := c.ShouldBind(&dataMsg); err != nil {
		log.Debug(f+"parsing failed: ", err)
		c.JSON(400, gin.H{"error": "Invalid API query"})
		return
	}
	log.WithFields(logrus.Fields{
		"Name":       dataMsg.Name,
		"Surname":    dataMsg.Surname,
		"Patronymic": dataMsg.Patronymic,
	}).Debug(f + "dataMsg")
	result := dataMsg.IsValid()
	if result != "" {
		log.Debug(f+"invalid message: ", result)
		dataMsg.Error = result
		c.JSON(422, gin.H{"error": dataMsg.Error})
		return
	}
	entry := models.Entry{
		Name:       dataMsg.Name,
		Surname:    dataMsg.Surname,
		Patronymic: dataMsg.Patronymic,
	}
	err := entry.Enrich(entry.Name)
	if err != nil {
		log.Error(f+"failed to enrich data from API: ", err)
		dataMsg.Error = fmt.Sprintf("Failed to enrich data from API: %v", err)
		c.JSON(500, gin.H{"error": dataMsg.Error})
		return
	}
	log.WithFields(logrus.Fields{
		"ID":          entry.ID,
		"Name":        entry.Name,
		"Surname":     entry.Surname,
		"Patronymic":  entry.Patronymic,
		"Age":         entry.Age,
		"Gender":      entry.Gender,
		"Nationality": entry.Nationality,
	}).Debug(f + "entry")
	err = entry.IsValid()
	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("Filling errors: %v", err)})
		return
	}
	err = db.C.Create(&entry).Error
	if err != nil {
		log.Error(f+"failed to create entry: ", err)
		c.JSON(500, gin.H{"error": "Failed to create entry"})
		return
	}
	c.JSON(200, gin.H{"message": "Success"})
}

// This API handler reads filtering parameters and get data from the
// database. Return a JSON message with data or an error with its
// cause.
func Read(c *gin.Context) {
	f := logging.F()
	pageSize := c.DefaultQuery("size", "10")
	pageNum := c.DefaultQuery("page", "1")
	filterCol := c.Query("col")
	filterData := c.Query("data")
	log.WithFields(logrus.Fields{
		"Size":   pageSize,
		"Num":    pageNum,
		"Column": filterCol,
		"Data":   filterData,
	}).Debug(f + "GET filters")
	switch {
	case filterCol != "" && filterData == "":
		fallthrough
	case filterCol == "" && filterData != "":
		c.JSON(400, gin.H{"error": `Fill in both "col" and "data"`})
		return
	}
	intSize, err := strconv.Atoi(pageSize)
	if err != nil {
		log.Debug(f+"invalid page size: ", err)
		c.JSON(400, gin.H{"error": "Invalid size parameter"})
		return
	}
	intPage, err := strconv.Atoi(pageNum)
	if err != nil {
		log.Debug(f+"invalid page number: ", err)
		c.JSON(400, gin.H{"error": "Invalid page parameter"})
		return
	}
	offset := (intPage - 1) * intSize
	var entries []models.Entry
	switch {
	case filterCol != "" && filterData != "":
		err = db.C.Model(&models.Entry{}).
			Limit(intSize).
			Offset(offset).
			Where(filterCol+" LIKE ?", "%"+filterData+"%").
			Find(&entries).
			Error
	default:
		err = db.C.Model(&models.Entry{}).
			Limit(intSize).
			Offset(offset).
			Find(&entries).
			Error
	}
	if err != nil {
		log.Error(f+"request to the database failed: ", err)
		c.JSON(500, gin.H{"error": "Request failed"})
		return
	}
	c.JSON(200, gin.H{"entries": entries})
}

// This API handler checks the input data, updates the record into the
// database. Return a JSON success message or an error with its cause.
func Update(c *gin.Context) {
	f := logging.F()
	var updEntry models.Entry
	if err := c.ShouldBind(&updEntry); err != nil {
		log.Debug(f+"parsing failed: ", err)
		c.JSON(400, gin.H{"error": "Invalid API query"})
		return
	}
	log.WithFields(logrus.Fields{
		"ID":          updEntry.ID,
		"Name":        updEntry.Name,
		"Surname":     updEntry.Surname,
		"Patronymic":  updEntry.Patronymic,
		"Age":         updEntry.Age,
		"Gender":      updEntry.Gender,
		"Nationality": updEntry.Nationality,
	}).Debug(f + "updEntry")
	err := updEntry.IsValid()
	if err != nil {
		c.JSON(422, gin.H{"error": fmt.Sprintf("Filling errors: %v", err)})
		return
	}
	err = db.C.Model(&models.Entry{}).
		Where("id = ?", updEntry.ID).
		Updates(map[string]interface{}{
			"name":        updEntry.Name,
			"surname":     updEntry.Surname,
			"patronymic":  updEntry.Patronymic,
			"age":         updEntry.Age,
			"gender":      updEntry.Gender,
			"nationality": updEntry.Nationality,
		}).
		Error
	if err != nil {
		c.JSON(
			404,
			gin.H{"message": fmt.Sprintf(
				`Entry "%v" does not exist`,
				updEntry.ID,
			)},
		)
		return
	}
	c.JSON(200, gin.H{"message": "Success"})
}

// This API handler checks the input ID, deletes the record from the
// database. Return a JSON success message or an error with its cause.
func Delete(c *gin.Context) {
	f := logging.F()
	var delEntry models.Entry
	if err := c.ShouldBind(&delEntry); err != nil {
		log.Debug(f+"parsing failed: ", err)
		c.JSON(400, gin.H{"error": "Invalid API query"})
		return
	}
	log.WithFields(logrus.Fields{
		"ID": delEntry.ID,
	}).Debug(f + "delEntry")
	var entry models.Entry
	err := db.C.First(&entry, "id = ?", delEntry.ID).Error
	if err != nil {
		c.JSON(
			404,
			gin.H{"message": fmt.Sprintf(
				`Entry "%v" does not exist`,
				delEntry.ID,
			)},
		)
		return
	}
	err = db.C.Unscoped().Delete(&entry).Error
	if err != nil {
		log.Error(f+"failed to delete entry: ", err)
		c.JSON(500, gin.H{"error": "Failed to delete entry"})
		return
	}
	c.JSON(200, gin.H{"message": "Success"})
}
