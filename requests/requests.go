package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// Gorutin for obtaining age data based on a name.
func Age(name string, age *uint8, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	url := fmt.Sprintf("https://api.agify.io/?name=%s", name)
	var reqData map[string]interface{}
	err := apiReq(url, &reqData)
	if err != nil {
		ch <- err
	}
	target, ok := reqData["age"].(float64) // int float64
	if !ok {
		ch <- errors.New("age data not found")
	}
	*age = uint8(target)
}

// Gorutin for obtaining gender data based on a name.
func Gender(name string, gender *string, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()
	url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
	var reqData map[string]interface{}
	err := apiReq(url, &reqData)
	if err != nil {
		ch <- err
	}
	target, ok := reqData["gender"].(string)
	if !ok {
		ch <- errors.New("gender data not found")
	}
	//time.Sleep(3 * time.Second)
	*gender = target
}

// Gorutin for obtaining nationality data based on a name.
func Nationality(
	name string, nation *string, wg *sync.WaitGroup, ch chan error,
) {
	defer wg.Done()
	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
	var reqData map[string]interface{}
	err := apiReq(url, &reqData)
	if err != nil {
		ch <- err
	}
	countryList, ok := reqData["country"].([]interface{})
	if !ok || len(countryList) == 0 {
		ch <- errors.New("country data not found")
	}
	firstCountry, ok := countryList[0].(map[string]interface{})
	if !ok {
		ch <- errors.New("invalid country data")
	}
	countryID, ok := firstCountry["country_id"].(string)
	if !ok {
		ch <- errors.New("country ID not found")
	}
	//time.Sleep(3 * time.Second)
	*nation = countryID
}

// The function of processing the request to the specified url. Fills
// out data map from the response body, otherwise returns an error.
func apiReq(url string, reqData *map[string]interface{}) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&reqData)
	if err != nil {
		return err
	}
	return nil
}
