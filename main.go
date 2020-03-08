package main

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/olahol/go-imageupload"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	vision "cloud.google.com/go/vision/apiv1"
)

func main() {
	gin.DefaultWriter = colorable.NewColorableStdout()
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./assets", true)))

	r.POST("/upload", uploadHandler)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "5000"
	}
	r.Run(":" + port)
}

func uploadHandler(c *gin.Context) {
	img, err := imageupload.Process(c.Request, "file")
	if err != nil {
		panic(err)
	}
	thumb, err := imageupload.ThumbnailPNG(img, 300, 300)
	if err != nil {
		panic(err)
	}
	h := sha1.Sum(thumb.Data)
	filepath := fmt.Sprintf(
		"%s_%x.png",
		time.Now().Format("20060102150405"), h[:4],
	)
	err = thumb.Save(filepath)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}
	detectDocumentText(filepath)
	os.Remove(filepath)
}

func detectDocumentText(filepath string) error {
	client, ctx := generateClient()
	defer client.Close()

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}
	text, err := client.DetectDocumentText(ctx, image, nil)
	log.Println(text.GetText())
	os.Remove(filepath)
	return nil
}

func generateClient() (*vision.ImageAnnotatorClient, context.Context) {
	json, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		log.Fatalln("Environment variable GOOGLE_APPLICATION_CREDENTIALS was not found.")
	}
	jwtConfig, err := google.JWTConfigFromJSON([]byte(json), vision.DefaultAuthScopes()...)
	if err != nil {
		panic(errors.New("google.JWTConfigFromJSON :" + err.Error() + "\n" + json))
	}
	ctx := context.Background()
	ts := jwtConfig.TokenSource(ctx)
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client, ctx
}
