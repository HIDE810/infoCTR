package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetLFCS(buf *bytes.Buffer) string {
	return fmt.Sprintf("%X", buf.Bytes()[0x110:0x118])
}

func GetKeyY(buf *bytes.Buffer) string {
	return fmt.Sprintf("%X", buf.Bytes()[0x110:0x120])
}

func GetID0(buf *bytes.Buffer) string {
	var id0 string
	var id0_array [4]uint32

	keyY, _ := hex.DecodeString(GetKeyY(buf))
	hash := sha256.Sum256(keyY)
	binary.Read(bytes.NewBuffer(hash[:16]), binary.LittleEndian, &id0_array)

	for _, v := range id0_array {
		id0 += fmt.Sprintf("%X", v)
	}
	return id0
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/assets", "./assets")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/upload", func(c *gin.Context) {

		status := "OK"
		buf := bytes.NewBuffer(nil)
		file, _, err := c.Request.FormFile("file")

		if err != nil {
			status = err.Error()
		} else {
			_, err = io.Copy(buf, file)
		}

		if err != nil {
			status = err.Error()
		}

		defer file.Close()

		if status != "OK" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"status": status,
			})
		} else if string(buf.Bytes()[0:4]) != "SEED" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"status": "error: invalid file",
			})
		} else {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"status": "OK",
				"lfcs":   GetLFCS(buf),
				"keyY":   GetKeyY(buf),
				"id0":    GetID0(buf),
			})
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
