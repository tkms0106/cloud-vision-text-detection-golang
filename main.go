package main

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/olahol/go-imageupload"
)

func main() {
	gin.DefaultWriter = colorable.NewColorableStdout()
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./assets", true)))

	r.POST("/upload", func(c *gin.Context) {
		img, err := imageupload.Process(c.Request, "file")
		if err != nil {
			panic(err)
		}
		thumb, err := imageupload.ThumbnailPNG(img, 300, 300)
		if err != nil {
			panic(err)
		}
		h := sha1.Sum(thumb.Data)
		thumb.Save(fmt.Sprintf("files/%s_%x.png",
			time.Now().Format("20060102150405"), h[:4]))
	})

	r.Run(":5000")
}
