package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
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

// Get the path to ChromeDriver based on the operating system
func getChromeDriverPath() string {
	switch runtime.GOOS {
	case "windows":
		return "./chromedriver-win64/chromedriver.exe"
	case "linux":
		return "./chromedriver-linux64/chromedriver"
	default:
		return ""
	}
}

// retry tries the given function up to maxRetries with a delay between retries
func retry(attempts int, delay time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			log.Printf("Attempt %d failed: %v", i+1, err)
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return fmt.Errorf("after %d attempts, the operation failed", attempts)
}

// Waits for an element with a timeout
func waitForElement(wd selenium.WebDriver, by, value string, timeout time.Duration) (selenium.WebElement, error) {
	for end := time.Now().Add(timeout); time.Now().Before(end); time.Sleep(500 * time.Millisecond) {
		el, err := wd.FindElement(by, value)
		if err == nil {
			return el, nil
		}
	}
	return nil, fmt.Errorf("element not found: %s %s", by, value)
}

// waitForElementWithRetry retries the fetching of an element with retries
func waitForElementWithRetry(wd selenium.WebDriver, by, value string, timeout time.Duration, attempts int, delay time.Duration) (selenium.WebElement, error) {
	var el selenium.WebElement
	err := retry(attempts, delay, func() error {
		var err error
		el, err = waitForElement(wd, by, value, timeout)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find element %s %s after retries: %v", by, value, err)
	}
	return el, nil
}

// Fetch data using Selenium with retry logic for stability
func fetchData() {
	seleniumPath := getChromeDriverPath()
	if seleniumPath == "" {
		log.Fatalf("Unsupported operating system: %v", runtime.GOOS)
	}
	const (
		port            = 9515
		chromeDriverURL = "http://localhost:%d/wd/hub"
		maxRetries      = 3
		retryDelay      = 2 * time.Second
	)

	opts := []selenium.ServiceOption{
		// Configure options here if needed
	}

	service, err := selenium.NewChromeDriverService(seleniumPath, port, opts...)
	if err != nil {
		log.Fatalf("Error starting the ChromeDriver server: %v", err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{
		"browserName": "chrome",
		"goog:chromeOptions": map[string]interface{}{
			"args": []string{
				"--headless",
				"--no-sandbox",
				"--disable-dev-shm-usage",
				"--disable-gpu",
			},
		},
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf(chromeDriverURL, port))
	if err != nil {
		log.Fatalf("Error connecting to WebDriver: %v", err)
	}
	defer wd.Quit()

	// Navigate to page
	err = retry(maxRetries, retryDelay, func() error {
		return wd.Get("https://www.simbio.si/sl/moj-dan-odvoza-odpadkov")
	})
	if err != nil {
		log.Fatalf("Failed to load page after retries: %v", err)
	}

	// Retry finding and interacting with the address input
	addressInput, err := waitForElementWithRetry(wd, selenium.ByCSSSelector, ".ui-comboBox-input", 10*time.Second, maxRetries, retryDelay)
	if err != nil {
		log.Fatalf("Error finding address input after retries: %v", err)
	}

	err = retry(maxRetries, retryDelay, func() error {
		return addressInput.SendKeys("ZAČRET 69,")
	})
	if err != nil {
		log.Fatalf("Error typing address after retries: %v", err)
	}

	// Retry finding and clicking the address suggestion
	addressSuggestion, err := waitForElementWithRetry(wd, selenium.ByXPATH, "//li[contains(text(), 'ZAČRET 69 , LJUBEČNA')]", 5*time.Second, maxRetries, retryDelay)
	if err != nil {
		log.Fatalf("Error finding address suggestion after retries: %v", err)
	}

	err = retry(maxRetries, retryDelay, func() error {
		return addressSuggestion.Click()
	})
	if err != nil {
		log.Fatalf("Error clicking address suggestion after retries: %v", err)
	}

	// Helper function to fetch data
	scrapeWasteData := func(labelSelector, dateSelector string) (string, string) {
		labelEl, _ := waitForElementWithRetry(wd, selenium.ByCSSSelector, labelSelector, 5*time.Second, maxRetries, retryDelay)
		dateEl, _ := waitForElementWithRetry(wd, selenium.ByCSSSelector, dateSelector, 5*time.Second, maxRetries, retryDelay)
		label, _ := labelEl.Text()
		date, _ := dateEl.Text()
		return label, date
	}

	// Scrape all waste categories
	mkoName, mkoDate := scrapeWasteData("div.next_mko > div.label", "div.next_mko > div.text")
	embName, embDate := scrapeWasteData("div.next_emb > div.label", "div.next_emb > div.text")
	bioName, bioDate := scrapeWasteData("div.next_bio > div.label", "div.next_bio > div.text")

	mutex.Lock()
	wasteData = struct {
		MKOName string
		MKODate string
		EmbName string
		EmbDate string
		BioName string
		BioDate string
	}{
		MKOName: mkoName,
		MKODate: mkoDate,
		EmbName: embName,
		EmbDate: embDate,
		BioName: bioName,
		BioDate: bioDate,
	}
	mutex.Unlock()
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

// Updates data every 15 minutes
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
