package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// filename holds the collection of image files to choose from
var filenames = []string{
	"Angrybird.JPG",
	"Arco&Tub.png",
	"IMG_9343.jpg",
	"angry-lemon-ufo.JPG",
	"austintiara4.png",
	"baby-geese.jpg",
	"bbq.jpg",
	"beach.JPG",
	"bunny-mask.jpg",
	"busted-light.jpg",
	"cat-glowing-eyes.JPG",
	"cat-on-leash.JPG",
	"cat.jpg",
	"clementine.png",
	"cow-peeking.jpg",
	"different-animals-01.png",
	"dratini.png",
	"everything-is-an-experiment.png",
	"experiment.png",
	"fine-food.jpg",
	"flower.jpg",
	"frenwho.png",
	"genshin-spa.jpg",
	"grass-and-desert-guy.png",
	"honeycomb-dogfood-logo.png",
	"horse-maybe.png",
	"is-this-emeri.png",
	"jean-and-statue.png",
	"jessitron.png",
	"keys-drying.jpg",
	"lime-on-soap-dispenser.jpg",
	"loki-closeup.jpg",
	"lynia.png",
	"ninguang-at-work.png",
	"paul-r-allen.png",
	"please.png",
	"roswell-nose.jpg",
	"roswell.JPG",
	"salt-packets-in-jar.jpg",
	"scarred-character.png",
	"square-leaf-with-nuts.jpg",
	"stu.jpeg",
	"sweating-it.png",
	"tanuki.png",
	"tennessee-sunset.JPG",
	"this-is-fine-trash.jpg",
	"three-pillars-2.png",
	"trash-flat.jpg",
	"walrus-painting.jpg",
	"windigo.png",
	"yellow-lines.JPG",
}

// ImageUrl is a struct to map the JSON output
type ImageUrl struct {
	ImageUrl string `json:"imageUrl"`
}

var bucketName string
var imageUrls []string

// init is special function that gets called before main
func init() {
	// Get the bucket name from the environment variable, with a default value
	bucketName = os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "random-pictures"
	}

	for _, filename := range filenames {
		url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, filename)
		imageUrls = append(imageUrls, url)
	}
}

func main() {

	// create a new echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/health", healthCheckHandler)

	// define a route '/imageUrl'
	e.GET("/imageUrl", imageUrlHandler)

	// start the server on the specified port
	e.Logger.Fatal(e.Start(":10116"))
}

func imageUrlHandler(c echo.Context) error {

	// select a random image url
	randomIndex := rand.Intn(len(imageUrls))
	selectedUrl := imageUrls[randomIndex]

	// create a image url struct with the selected image url
	response := ImageUrl{ImageUrl: selectedUrl}

	// return the response
	return c.JSON(http.StatusOK, response)
}

func healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}
