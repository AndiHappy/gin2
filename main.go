package main

import (
	"fmt"
	"gin2/gin"
	"net/http"
)

func main() {
	// This is a placeholder for the main function.
	// You can add your application logic here.
	println("Hello, World!")
	// Example of a simple function call
	gin.SetMode(gin.DebugMode)
	fmt.Println("Hello Gin")
	router := gin.Default()
	router.GET("/a", func(c *gin.Context) {
		data := map[string]interface{}{"/a": "1"}
		c.JSON(http.StatusOK, data)
	})
	router.GET("/a/b", func(c *gin.Context) {
		data := map[string]interface{}{"/a/b": "1"}
		c.JSON(http.StatusOK, data)
	})
	router.GET("/a/b/c", func(c *gin.Context) {
		data := map[string]interface{}{"/a/b/c": "1"}
		c.JSON(http.StatusOK, data)
	})
	err := router.Run(":8080")
	if err != nil {
		return
	}

}
