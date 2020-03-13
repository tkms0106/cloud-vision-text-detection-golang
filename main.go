package main

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/olahol/go-imageupload"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	vision "cloud.google.com/go/vision/apiv1"
)

func main() {
	client := generateClient()
	defer client.Close()

	gin.DefaultWriter = colorable.NewColorableStdout()
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./assets", true)))
	r.POST("/upload", uploadHandlerFunc(client))

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}

	err := r.Run(":" + port)
	if err != nil {
		panic(err)
	}
}

func uploadHandlerFunc(client *vision.ImageAnnotatorClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		size, err := humanize.ParseBytes("10MB")
		if err != nil {
			log.Fatalf("%v", err)
		}

		imageupload.LimitFileSize(int64(size), c.Writer, c.Request)
		img, err := imageupload.Process(c.Request, "file")
		if err != nil {
			panic(err)
		}

		h := sha1.Sum(img.Data)
		filepath := fmt.Sprintf(
			"%s_%x.png",
			time.Now().Format("20060102150405"), h[:4],
		)
		err = img.Save(filepath)
		if err != nil {
			panic(err)
		}

		text := detectDocumentText(client, filepath)
		c.JSON(200, struct {
			Text string `json:"text"`
		}{text})
		os.Remove(filepath)
	}
}

func detectDocumentText(client *vision.ImageAnnotatorClient, filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	ctx := context.Background()
	text, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		log.Fatalf("Failed to detect document text: %v", err)
	}

	log.Printf("%v", text.GetText())
	return text.GetText()
}

func generateClient() *vision.ImageAnnotatorClient {
	json, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		log.Fatalf("Environment variable GOOGLE_APPLICATION_CREDENTIALS was not found.")
	}

	jwtConfig, err := google.JWTConfigFromJSON([]byte(json), vision.DefaultAuthScopes()...)
	if err != nil {
		log.Fatalf("%v", errors.New("google.JWTConfigFromJSON :"+err.Error()+"\n"+json))
	}

	ctx := context.Background()
	ts := jwtConfig.TokenSource(ctx)
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return client
}
