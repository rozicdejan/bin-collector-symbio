package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/tebeka/selenium"
)

var (
	wasteData = struct {
		MKOName string
		MKODate string
		EmbName string
		EmbDate string
		BioName string
		BioDate string
	}{}
	mutex = &sync.Mutex{}
)

// Fetch data using Selenium
func fetchData() {
	const (
		seleniumPath    = ".\\chromedriver-win64\\chromedriver.exe"
		port            = 9515
		chromeDriverURL = "http://localhost:%d/wd/hub"
	)

	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(seleniumPath, port, opts...)
	if err != nil {
		log.Fatalf("Error starting the ChromeDriver server: %v", err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf(chromeDriverURL, port))
	if err != nil {
		log.Fatalf("Error connecting to WebDriver: %v", err)
	}
	defer wd.Quit()

	err = wd.Get("https://www.simbio.si/sl/moj-dan-odvoza-odpadkov")
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}
	time.Sleep(3 * time.Second)

	addressInput, err := wd.FindElement(selenium.ByCSSSelector, ".ui-comboBox-input")
	if err != nil {
		log.Fatalf("Could not find address input: %v", err)
	}
	err = addressInput.SendKeys("ZAČRET 69,")
	if err != nil {
		log.Fatalf("Error typing address: %v", err)
	}
	time.Sleep(2 * time.Second)

	addressSuggestion, err := wd.FindElement(selenium.ByXPATH, "//li[contains(text(), 'ZAČRET 69 , LJUBEČNA')]")
	if err != nil {
		log.Fatalf("Error finding address suggestion: %v", err)
	}
	err = addressSuggestion.Click()
	if err != nil {
		log.Fatalf("Error clicking address suggestion: %v", err)
	}
	time.Sleep(5 * time.Second)

	// Scrape data
	nameElement, _ := wd.FindElement(selenium.ByCSSSelector, "div.next_mko > div.label")
	dateElement, _ := wd.FindElement(selenium.ByCSSSelector, "div.next_mko > div.text")
	wasteData.MKOName, _ = nameElement.Text()
	wasteData.MKODate, _ = dateElement.Text()

	nameElement, _ = wd.FindElement(selenium.ByCSSSelector, "div.next_emb > div.label")
	dateElement, _ = wd.FindElement(selenium.ByCSSSelector, "div.next_emb > div.text")
	wasteData.EmbName, _ = nameElement.Text()
	wasteData.EmbDate, _ = dateElement.Text()

	nameElement, _ = wd.FindElement(selenium.ByCSSSelector, "div.next_bio > div.label")
	dateElement, _ = wd.FindElement(selenium.ByCSSSelector, "div.next_bio > div.text")
	wasteData.BioName, _ = nameElement.Text()
	wasteData.BioDate, _ = dateElement.Text()

	mutex.Lock()
	defer mutex.Unlock()
}

// Serve dynamic HTML
func dataHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, wasteData)
}

// Updates data every 15 seconds
func dataUpdater() {
	for {
		fetchData()
		time.Sleep(15 * time.Minute)
	}
}

func main() {
	// Serve static files (for SVG images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	go dataUpdater() // Runs the data updater in the background

	http.HandleFunc("/", dataHandler)
	// Change from localhost to 0.0.0.0 to bind to all network interfaces
	fmt.Println("Server running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
