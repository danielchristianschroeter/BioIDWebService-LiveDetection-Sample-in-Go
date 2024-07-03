package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Used for build version information
var version = "development"

// AppConfig holds the configuration for the application
type AppConfig struct {
	BWSAppID         string
	BWSAppSecret     string
	Image1           string
	Image2           string
	DetailedResponse bool
}

// Response represents the JSON response structure
type Response struct {
	Success bool   `json:"Success"`
	State   string `json:"State"`
	JobID   string `json:"JobID"`
	Samples []struct {
		Errors []struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
			Details string `json:"Details"`
		} `json:"Errors"`
		EyeCenters struct {
			RightEyeX float64 `json:"RightEyeX"`
			RightEyeY float64 `json:"RightEyeY"`
			LeftEyeX  float64 `json:"LeftEyeX"`
			LeftEyeY  float64 `json:"LeftEyeY"`
		} `json:"EyeCenters"`
	} `json:"Samples"`
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Println("BioIDWebService LiveDetection Sample in Go. Version: " + version)
		flag.PrintDefaults()
	}
	
	config := parseFlags()
	client := &http.Client{Timeout: 30 * time.Second}

	base64Image1 := imageToBase64(config.Image1)
	base64Image2 := imageToBase64(config.Image2)

	statusCode, response := sendRequest(client, http.MethodPost, config, base64Image1, base64Image2)
	processResponse(statusCode, response, config.DetailedResponse)
}

// parses the command-line flags and returns an AppConfig struct
func parseFlags() AppConfig {
	var config AppConfig
	flag.StringVar(&config.BWSAppID, "BWSAppID", "", "BioIDWebService AppID")
	flag.StringVar(&config.BWSAppSecret, "BWSAppSecret", "", "BioIDWebService AppSecret")
	flag.StringVar(&config.Image1, "image1", "", "1st source image")
	flag.StringVar(&config.Image2, "image2", "", "2nd source image")
	flag.BoolVar(&config.DetailedResponse, "detailedResponse", false, "Return detailed JSON output of response")
	flag.Parse()

	if config.BWSAppID == "" || config.BWSAppSecret == "" || config.Image1 == "" || config.Image2 == "" {
		log.Fatal("Usage: -BWSAppID <BWSAppID> -BWSAppSecret <BWSAppSecret> -image1 <image1> -image2 <image2>")
	}
	return config
}

// Sends an HTTP request and returns the status code and response body
func sendRequest(client *http.Client, method string, config AppConfig, liveimage1, liveimage2 string) (int, []byte) {
	endpoint := "https://bws.bioid.com/extension/livedetection"
	if config.DetailedResponse {
		endpoint += "?state=true"
	}

	values, _ := json.Marshal(map[string]string{
		"liveimage1": liveimage1,
		"liveimage2": liveimage2,
	})

	req, err := http.NewRequest(method, endpoint, strings.NewReader(string(values)))
	if err != nil {
		log.Fatalf("An error occurred %v", err)
	}

	req.SetBasicAuth(config.BWSAppID, config.BWSAppSecret)
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Could not parse response body. %+v", err)
	}

	return response.StatusCode, body
}

// processes the HTTP response, printing out detailed information if required
func processResponse(statusCode int, response []byte, detailed bool) {
	if statusCode == http.StatusOK {
		if detailed {
			var result Response
			if err := json.Unmarshal(response, &result); err != nil {
				log.Println("Can not unmarshal JSON.")
			}
			log.Println("Detailed response body:\n" + prettyPrint(result))
			if result.Success {
				log.Println("Result:\nImages are recorded from a live person.")
			} else {
				log.Println("Result:\nImages are NOT recorded from a live person.")
			}
		} else {
			log.Println(string(response))
		}
	} else {
		log.Fatalf("Received http response code != 200: %v", statusCode)
	}
}

// Converts an image file to a base64-encoded string
func imageToBase64(file string) string {
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var builder strings.Builder
	mimeType := http.DetectContentType(bytes)

	switch mimeType {
	case "image/jpeg":
		builder.WriteString("data:image/jpeg;base64,")
	case "image/png":
		builder.WriteString("data:image/png;base64,")
	default:
		log.Fatalf("mime type %v for %v is not supported.", mimeType, file)
	}

	builder.WriteString(base64.StdEncoding.EncodeToString(bytes))
	return builder.String()
}

// returns a pretty-printed JSON representation of an object
func prettyPrint(i interface{}) string {
	s, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		log.Fatalf("JSON marshal error: %v", err)
	}
	return string(s)
}