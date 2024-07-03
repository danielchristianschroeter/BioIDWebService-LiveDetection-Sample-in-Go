# BioIDWebService-LiveDetection-Sample-in-Go

This repository contains a very simple example command line implementation of the **[LiveDetection REST API of the BioID WebService](https://developer.bioid.com/classicbws/bwsreference/webapi/livedetection)** in Go.

Two recorded images are required to perform a face liveness detection.

## Requirements

Before you can use the LiveDetection API, you need to create a BWS App identifier and App secret in the BWS Portal.
You can request trial access on https://bwsportal.bioid.com/register

## Usage

1. Build or download prebuild executable
2. Execute following command to perform a face liveness detection with two images:

```
./BioIDWebService-LiveDetection-Sample-in-Go -BWSAppID <BWSAppID> -BWSAppSecret <BWSAppSecret> -image1 ./example_images/live_image1_without_errors.jpg -image2 ./example_images/live_image2_without_errors.jpg
```

You can use the example test images inside the example_images folder:
fake_image1,2_with_errors.jpg => LiveDetectionFailed, ImageOverExposure
fake_image1,2_without_errors.jpg => LiveDetectionFailed
live_image1,2_with_errors.jpg => ImageOverExposure
live_image1,2_without_errors.jpg

Example Output:

`true` or `false`

### Available command line parameter

```
./BioIDWebService-LiveDetection-Sample-in-Go --help
  -BWSAppID string
    	BioIDWebService AppID
  -BWSAppSecret string
    	BioIDWebService AppSecret
  -detailedResponse
    	Return detailed JSON output of response
  -image1 string
    	1st source image
  -image2 string
    	2nd source image
```

## Clone and build the project

```
$ git clone https://github.com/danielchristianschroeter/BioIDWebService-LiveDetection-Sample-in-Go
$ cd BioIDWebService-LiveDetection-Sample-in-Go
$ go build .
```
