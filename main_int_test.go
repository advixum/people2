package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	db "people2/database"
	"people2/models"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

// Requirements: .env PostgreSQL credentials

// Testing data processing in the handlers.Create() function.
func TestCreateAPI(t *testing.T) {
	type args struct {
		data  models.FullName
		valid bool
	}
	tests := []struct {
		test string
		args args
	}{
		{
			test: "Valid data was saved and enriched",
			args: args{
				valid: true,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Valid data with empty patronymic was saved and enriched",
			args: args{
				valid: true,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "Ivanov",
					Patronymic: "",
				},
			},
		},
		{
			test: "Valid data without patronymic was saved and enriched",
			args: args{
				valid: true,
				data: models.FullName{
					Name:    "Ivan",
					Surname: "Ivanov",
				},
			},
		},
		{
			test: "Empty name was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "",
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Data without name was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Less than 2 letters name was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "N",
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "More than 50 letters name was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name: `
						Nnnnnnnnnn
						Nnnnnnnnnn
						Nnnnnnnnnn
						Nnnnnnnnnn
						NnnnnnnnnnN
					`,
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Name with numbers was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "1Ivan",
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Name with symbols was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "!Ivan",
					Surname:    "Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Empty surname was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Data without surname was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "Ivan",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Less than 2 letters surname was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "S",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "More than 50 letters surname was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name: "Ivan",
					Surname: `
						Nnnnnnnnnn
						Nnnnnnnnnn
						Nnnnnnnnnn
						Nnnnnnnnnn
						NnnnnnnnnnN
					`,
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Surname with numbers was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "1Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
		{
			test: "Surname with symbols was rejected",
			args: args{
				valid: false,
				data: models.FullName{
					Name:       "Ivan",
					Surname:    "!Ivanov",
					Patronymic: "Ivanovich",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.Entry{})
			defer db.C.Migrator().DropTable(&models.Entry{})

			// Create testing data
			send := tt.args.data
			jsonData, err := json.Marshal(send)
			assert.NoError(t, err)

			// Setup router
			r := router()
			request, err := http.NewRequest(
				"POST",
				"http://127.0.0.1:8080/api/create",
				bytes.NewBuffer(jsonData),
			)
			assert.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Get database values
			var entry models.Entry
			err = db.C.First(&entry).Error

			// Estimation of values
			if tt.args.valid {
				assert.Equal(t, 200, response.Code)
				assert.NoError(t, err)
			} else {
				assert.NotEqual(t, 200, response.Code)
				assert.Error(t, err)
			}
		})
	}
}

// Testing data processing in the handlers.Read() function.
func TestReadAPI(t *testing.T) {
	type args struct {
		valid   bool
		size    int
		page    int
		col     string
		data    string
		entries []models.Entry
	}
	tests := []struct {
		test string
		args args
	}{
		{
			test: "The entries list with 3 records was return",
			args: args{
				valid: true,
				entries: []models.Entry{
					{
						Name:        "Ivan",
						Surname:     "Ivanov",
						Patronymic:  "Ivanovich",
						Age:         42,
						Gender:      "male",
						Nationality: "RU",
					},
					{
						Name:        "Anna",
						Surname:     "Ivanova",
						Patronymic:  "Ivanovna",
						Age:         42,
						Gender:      "female",
						Nationality: "RU",
					},
					{
						Name:        "Ivan",
						Surname:     "Ushakov",
						Patronymic:  "Vasilevich",
						Age:         30,
						Gender:      "male",
						Nationality: "RU",
					},
				},
			},
		},
		{
			test: "The empty entries list was return",
			args: args{
				valid:   true,
				entries: []models.Entry{},
			},
		},
		{
			test: "Valid paginated data was return",
			args: args{
				valid: true,
				size:  1,
				page:  2,
				entries: []models.Entry{
					{
						Name:        "Ivan",
						Surname:     "Ivanov",
						Patronymic:  "Ivanovich",
						Age:         42,
						Gender:      "male",
						Nationality: "RU",
					},
					{
						Name:        "Anna",
						Surname:     "Ivanova",
						Patronymic:  "Ivanovna",
						Age:         42,
						Gender:      "female",
						Nationality: "RU",
					},
					{
						Name:        "Ivan",
						Surname:     "Ushakov",
						Patronymic:  "Vasilevich",
						Age:         30,
						Gender:      "male",
						Nationality: "RU",
					},
				},
			},
		},
		{
			test: "Valid filtrated data was return",
			args: args{
				valid: true,
				col:   "Name",
				data:  "Ivan",
				entries: []models.Entry{
					{
						Name:        "Ivan",
						Surname:     "Ivanov",
						Patronymic:  "Ivanovich",
						Age:         42,
						Gender:      "male",
						Nationality: "RU",
					},
					{
						Name:        "Anna",
						Surname:     "Ivanova",
						Patronymic:  "Ivanovna",
						Age:         42,
						Gender:      "female",
						Nationality: "RU",
					},
					{
						Name:        "Ivan",
						Surname:     "Ushakov",
						Patronymic:  "Vasilevich",
						Age:         30,
						Gender:      "male",
						Nationality: "RU",
					},
				},
			},
		},
		{
			test: "Filtration request without column was aborted",
			args: args{
				valid: false,
				col:   "",
				data:  "Ivan",
				entries: []models.Entry{
					{
						Name:        "Ivan",
						Surname:     "Ivanov",
						Patronymic:  "Ivanovich",
						Age:         42,
						Gender:      "male",
						Nationality: "RU",
					},
					{
						Name:        "Anna",
						Surname:     "Ivanova",
						Patronymic:  "Ivanovna",
						Age:         42,
						Gender:      "female",
						Nationality: "RU",
					},
					{
						Name:        "Ivan",
						Surname:     "Ushakov",
						Patronymic:  "Vasilevich",
						Age:         30,
						Gender:      "male",
						Nationality: "RU",
					},
				},
			},
		},
		{
			test: "Filtration request without data was aborted",
			args: args{
				valid: false,
				col:   "Name",
				data:  "",
				entries: []models.Entry{
					{
						Name:        "Ivan",
						Surname:     "Ivanov",
						Patronymic:  "Ivanovich",
						Age:         42,
						Gender:      "male",
						Nationality: "RU",
					},
					{
						Name:        "Anna",
						Surname:     "Ivanova",
						Patronymic:  "Ivanovna",
						Age:         42,
						Gender:      "female",
						Nationality: "RU",
					},
					{
						Name:        "Ivan",
						Surname:     "Ushakov",
						Patronymic:  "Vasilevich",
						Age:         30,
						Gender:      "male",
						Nationality: "RU",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.Entry{})
			defer db.C.Migrator().DropTable(&models.Entry{})

			// Create testing data
			db.C.Create(&tt.args.entries)

			// Setup router
			r := router()
			url := ""
			var pagination []string
			intSize := 10
			intPage := 1
			if tt.args.size != 0 {
				pagination = append(
					pagination,
					fmt.Sprintf("size=%v", tt.args.size),
				)
				intSize = tt.args.size
			}
			if tt.args.page != 0 {
				pagination = append(
					pagination,
					fmt.Sprintf("page=%v", tt.args.page),
				)
				intPage = tt.args.page
			}
			if tt.args.col != "" {
				pagination = append(pagination, "col="+tt.args.col)
			}
			if tt.args.data != "" {
				pagination = append(pagination, "data="+tt.args.data)
			}
			if len(pagination) == 0 {
				url = "http://127.0.0.1:8080/api/read"
			} else {
				params := strings.Join(pagination, "&")
				url = "http://127.0.0.1:8080/api/read?" + params
			}
			request, err := http.NewRequest(
				"GET",
				url,
				nil,
			)
			assert.NoError(t, err)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Get database values
			offset := (intPage - 1) * intSize
			var entries []models.Entry
			switch {
			case tt.args.col != "" && tt.args.data != "":
				err = db.C.Model(&models.Entry{}).
					Limit(intSize).
					Offset(offset).
					Where(tt.args.col+" LIKE ?", "%"+tt.args.data+"%").
					Find(&entries).
					Error
			default:
				err = db.C.Model(&models.Entry{}).
					Limit(intSize).
					Offset(offset).
					Find(&entries).
					Error
			}
			assert.NoError(t, err)
			entriesJSON, err := json.Marshal(gin.H{"entries": entries})
			assert.NoError(t, err)

			// Estimation of values
			if tt.args.valid {
				assert.Equal(t, 200, response.Code)
				assert.JSONEq(
					t,
					string(entriesJSON),
					strings.TrimSpace(response.Body.String()),
				)
			} else {
				assert.Equal(t, 400, response.Code)
				assert.NotEqual(
					t,
					string(entriesJSON),
					strings.TrimSpace(response.Body.String()),
				)
			}
		})
	}
}

// Testing data processing in the handlers.Update() function.
func TestUpdateAPI(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.Entry{})
	defer db.C.Migrator().DropTable(&models.Entry{})
	data := models.Entry{
		Name:        "Ivan",
		Surname:     "Ivanov",
		Patronymic:  "Ivanovich",
		Age:         42,
		Gender:      "male",
		Nationality: "RU",
	}
	err := db.C.Create(&data).Error
	assert.NoError(t, err)

	// Create testing data
	send := models.Entry{
		ID:          1,
		Name:        "Ivan",
		Surname:     "Smirnov",
		Patronymic:  "Ivanovich",
		Age:         42,
		Gender:      "male",
		Nationality: "RU",
	}
	jsonData, err := json.Marshal(send)
	assert.NoError(t, err)

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"PATCH",
		"http://127.0.0.1:8080/api/update",
		bytes.NewBuffer(jsonData),
	)
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get database values
	var entry models.Entry
	err = db.C.Where("name = ?", data.Name).First(&entry).Error

	// Estimation of values
	assert.Equal(t, 200, response.Code)
	assert.NoError(t, err)
	assert.Equal(t, send.Surname, entry.Surname)
}

// Testing data processing in the handlers.Delete() function.
func TestDeleteAPI(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.Entry{})
	defer db.C.Migrator().DropTable(&models.Entry{})
	data := models.Entry{
		Name:        "Ivan",
		Surname:     "Ivanov",
		Patronymic:  "Ivanovich",
		Age:         42,
		Gender:      "male",
		Nationality: "RU",
	}
	err := db.C.Create(&data).Error
	assert.NoError(t, err)

	// Create testing data
	send := models.Entry{
		ID: 1,
	}
	jsonData, err := json.Marshal(send)
	assert.NoError(t, err)

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"DELETE",
		"http://127.0.0.1:8080/api/delete",
		bytes.NewBuffer(jsonData),
	)
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get database values
	var entries []models.Entry
	err = db.C.Find(&entries).Error
	assert.NoError(t, err)
	entriesJSON, err := json.Marshal(gin.H{"entries": entries})
	assert.NoError(t, err)

	// Estimation of values
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, string(entriesJSON), "{\"entries\":[]}")
}
