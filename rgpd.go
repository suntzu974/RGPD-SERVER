package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
)

func main() {

	var err error
	var config = Configuration{}

	switch os := runtime.GOOS; os {
	case "darwin":
		log.Printf("Platform from configuration file   %s.\n", os)
		config = LoadConfiguration("/Users/suntzu974/GoProjects/RGPD-SERVER/rgpd.json")
	case "linux":
		log.Printf("Platform from configuration file   %s.\n", os)
		config = LoadConfiguration("/home/jeannick/PROJECTS/RFPD-SERVER/rgpd.json")
	default:
		log.Printf("Platform from configuration file   %s.\n", os)
		config = LoadConfiguration("rgpd.json")
	}

	http.HandleFunc("/", config.homeHandler)
	http.HandleFunc("/server", config.ServerHandle)
	http.HandleFunc("/customer", config.CustomerHandle)
	http.HandleFunc("/consent", config.ConsentHandle)
	http.HandleFunc("/consents", config.ConsentsHandle)
	http.HandleFunc("/sofarem", config.LoadSofarem)

	openLogFile(config.Log)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Server started at port %d\n", config.Port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

}
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
func openLogFile(logfile string) {
	if logfile != "" {
		lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

		if err != nil {
			log.Fatal("OpenLogfile: os.OpenFile:", err)
		}

		log.SetOutput(lf)
	}
}
