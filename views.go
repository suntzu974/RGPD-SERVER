package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (c *Configuration) ServerHandle(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
	io.WriteString(w, SelectVersion(GetDatabase(c)))
	io.WriteString(w, "\n")
}

func (c *Configuration) CustomerHandle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		query := req.URL.Query().Get("siret")
		customer, count, err := ReadCustomer(GetDatabase(c), query)
		if err != nil {
			log.Printf("Error Reading customer: %s ", err.Error())
		}
		fmt.Printf("Read %d rows successfully.\n", count)
		respondWithJSON(w, http.StatusOK, customer)
	case "POST":
		var customer Customer
		err := json.NewDecoder(req.Body).Decode(&customer)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error()+" Invalid request payload")
			return
		}
		defer req.Body.Close()

		_, erreur := CreateCustomer(GetDatabase(c), customer)
		if erreur != nil {
			log.Printf("Error Creating customer: %s ", erreur.Error())
		}
		respondWithJSON(w, http.StatusOK, customer)

	default:
		fmt.Fprintf(w, "GET or POST are accepted !")
	}
}
func (c *Configuration) ConsentsHandle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		consents, count, err := AllConsents(GetDatabase(c))
		if err != nil {
			log.Printf("Error Reading Configuration : %s ", err.Error())
		}
		fmt.Printf("Read %d rows successfully.\n", count)
		respondWithJSON(w, http.StatusOK, consents)
	default:
		fmt.Fprintf(w, "GET are accepted !")
	}
}
func (c *Configuration) ConsentHandle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		query := req.URL.Query().Get("siret")
		customer, count, err := ReadConsent(GetDatabase(c), query)
		if err != nil {
			log.Printf("Error Reading Consent: %s ", err.Error())
		}
		fmt.Printf("Read %d rows successfully.\n", count)
		respondWithJSON(w, http.StatusOK, customer)
	case "POST":
		var consent Consent
		//	c := M.Consent{Siret: "123456", UsingGeneralConditions: true, Newsletters: true, CommercialOffersByMail: true, CommercialOffersBySms: true, CommercialOffersByPost: true, Signature: "signature", CreatedAt: time.Now()}
		err := json.NewDecoder(req.Body).Decode(&consent)
		fmt.Println("Siret : ", consent.Siret)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer req.Body.Close()

		_, erreur := CreateConsent(GetDatabase(c), consent)
		if erreur != nil {
			log.Printf("Error Creating consent: %s", erreur.Error())
		}
		respondWithJSON(w, http.StatusOK, consent)
	case "PUT":
		var consent Consent
		err := json.NewDecoder(req.Body).Decode(&consent)
		fmt.Println("Siret : ", consent.Siret)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer req.Body.Close()

		_, erreur := UpdateConsent(GetDatabase(c), consent)
		if erreur != nil {
			log.Fatal("Error updating consent: " + erreur.Error())
		}
		respondWithJSON(w, http.StatusOK, consent)
	default:
		fmt.Fprintf(w, "GET,UPDATE or POST are accepted !")
	}
}
func (c *Configuration) LoadSofarem(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		stocks_sofarem, stocks_hometech, count, err := ReadStockFromSofarem(GetDatabase(c))
		if err != nil {
			log.Printf("Error Reading Stocks : %s ", err.Error())
		}
		fmt.Printf("Read %d rows successfully.\n", count)
		entity := []string{"SOFAREM", "HOMETECH"}
		xlsx := excelize.NewFile()
		WriteDataForEntity(xlsx, stocks_sofarem, entity[0])
		WriteDataForEntity(xlsx, stocks_hometech, entity[1])
		xlsx.DeleteSheet("Sheet1")

		err = xlsx.SaveAs("/tmp/SofaremStock.xlsx")
		if err != nil {
			fmt.Println(err)
		}
		Filename := "/tmp/SofaremStock.xlsx"
		if Filename == "" {
			http.Error(w, "Get 'file' not specified in url.", 400)
			return
		}

		//Check if file exists and open
		Openfile, err := os.Open(Filename)
		defer Openfile.Close() //Close after function return
		if err != nil {
			//File not found, send 404
			http.Error(w, "File not found.", 404)
			return
		}
		FileHeader := make([]byte, 512)
		Openfile.Read(FileHeader)
		FileContentType := http.DetectContentType(FileHeader)
		FileStat, _ := Openfile.Stat()                     //Get info from file
		FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
		w.Header().Set("Content-Disposition", "attachment; filename="+Filename)
		w.Header().Set("Content-Type", FileContentType)
		w.Header().Set("Content-Length", FileSize)
		Openfile.Seek(0, 0)
		io.Copy(w, Openfile)

	default:
		fmt.Fprintf(w, "GET are accepted !")
	}
}
func FixedLengthString(length int, str string) string {
	verb := fmt.Sprintf("%%%d.%ds", length, length)
	return fmt.Sprintf(verb, str)
}
func WriteDataForEntity(xlsx *excelize.File, stocks []Stock, entity string) {
	/* Header */
	sheet := xlsx.NewSheet(entity)
	stock := Stock{}
	e := reflect.ValueOf(&stock).Elem()
	cell := []string{"A1", "B1", "C1", "D1", "E1"}
	for i := 0; i < e.NumField(); i++ {
		xlsx.SetCellValue(entity, cell[i], e.Type().Field(i).Name)
	}
	// Data
	cell = []string{"A", "B", "C", "D", "E"}
	for index, data := range stocks {
		e := reflect.ValueOf(&data).Elem()
		for i := 0; i < e.NumField(); i++ {
			xlsx.SetCellValue(entity, cell[i]+strconv.Itoa(index+2), e.Field(i))
		}
	}
	xlsx.SetActiveSheet(sheet)
}
