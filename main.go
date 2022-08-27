package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Used for build version information
var version = "development"

// Global variables used for command line flags
var (
	BWSAppID         string
	BWSAppSecret     string
	image1           string
	image2           string
	detailedResponse bool
)

// Struct from the Response
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

func httpClient() *http.Client {
	client := &http.Client{Timeout: 30 * time.Second} // 30 seconds timeout
	return client
}

// BasicAuth required for BWS AppID and AppSecret
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Prepare http request
func sendRequest(client *http.Client, method string, BWSAppID string, BWSAppSecret string, liveimage1 string, liveimage2 string) (int, []byte) {
	endpoint := "https://bws.bioid.com/extension/livedetection"
	if detailedResponse {
		endpoint += "?state=true"
	}
	//endpoint := BWSEndpoint
	values, _ := json.Marshal(map[string]string{
		"liveimage1": liveimage1,
		"liveimage2": liveimage2,
	})
	// Using http.Request to make a request
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(values))
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Authorization", "Basic "+basicAuth(BWSAppID, BWSAppSecret))
	// Handle error
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
	}

	httpresponsestatuscode := response.StatusCode
	// Close the connection to reuse it
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Could not parse response body. %+v", err)
	}

	return httpresponsestatuscode, body
}

func imageToBase64(file string) string {
	// Read the entire file into a byte slice
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	var base64Encoding string
	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	default:
		log.Fatalf("mime type %v for %v is not supported.", mimeType, file)
	}

	// Append the base64 encoded output to image encoding
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	// Return the full base64 representation of the image
	return base64Encoding
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func init() {
	// Initialize command line flags
	flag.StringVar(&BWSAppID, "BWSAppID", "", "BioIDWebService AppID")
	flag.StringVar(&BWSAppSecret, "BWSAppSecret", "", "BioIDWebService AppSecret")
	flag.StringVar(&image1, "image1", "", "1st source image")
	flag.StringVar(&image2, "image2", "", "2nd source image")
	flag.BoolVar(&detailedResponse, "detailedResponse", false, "Return detailed JSON output of response")
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Println("BioIDWebService LiveDetection Sample in Go. Version: " + version)
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(BWSAppID) == 0 || len(BWSAppSecret) == 0 || len(image1) == 0 || len(image2) == 0 {
		log.Fatal("Usage: -BWSAppID <BWSAppID> -BWSAppSecret <BWSAppSecret> -image1 <image1> -image2 <image2>")
	}

	// Convert images to base64 string
	base64_image1 := imageToBase64(image1)
	base64_image2 := imageToBase64(image2)

	// c should be re-used for further calls
	c := httpClient()
	httpresponsestatuscode, response := sendRequest(c, http.MethodPost, BWSAppID, BWSAppSecret, base64_image1, base64_image2)

	// Validate http response code
	if httpresponsestatuscode == 200 {
		// detailed response output
		if detailedResponse {
			// Parse []byte to go struct pointer, required for PrettyPrint
			var result Response
			if err := json.Unmarshal(response, &result); err != nil {
				log.Println("Can not unmarshal JSON.")
			}
			log.Println("Detailed response body:\n" + PrettyPrint(result))
			if result.Success {
				log.Println("Result:\nImages are recorded from a live person.")
			} else {
				log.Println("Result:\nImages are NOT recorded from a live person.")
			}
			// Check for errors
			for i, samples := range result.Samples {
				// If Samples.Errors is not empty, return error
				if len(samples.Errors) > 0 {
					no := strconv.Itoa(i + 1)
					log.Println("Errors found for image" + no + ":")
					for _, errors := range samples.Errors {
						log.Println(errors.Code, "-", errors.Message, "-", errors.Details)
					}
				}
			}
		} else {
			// Default response
			log.Println(string(response))
		}
	} else {
		log.Fatalf("Received http response code != 200: %v", httpresponsestatuscode)
	}
}
