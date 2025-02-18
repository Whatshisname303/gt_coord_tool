package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Db struct {
	Users []User `json:"users"`
}

type User struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Paths []Path `json:"paths"`
}

type Node struct {
	Id        int    `json:"id"`
	Longitude int    `json:"longitude"`
	Latitude  int    `json:"latitude"`
	Name      string `json:"name,omitempty"`
}

type Path struct {
	Id   int `json:"id"`
	From int `json:"from"`
	To   int `json:"to"`
}

func readDb() (Db, error) {
	jsonFile, err := os.Open("db.json")
	if err != nil {
		return Db{}, err
	}
	defer jsonFile.Close()

	byteValues, _ := io.ReadAll(jsonFile)

	var data Db

	json.Unmarshal(byteValues, &data)
	return data, nil
}

func saveDb(db any) error {
	jsonString, _ := json.Marshal(db)
	os.WriteFile("db.json", jsonString, os.ModePerm)
	return nil
}

func main() {
	router := gin.Default()

	router.StaticFile("/", "./index.html")
	router.StaticFile("db", "./db.json")

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.GET("/fetch/:user", func(c *gin.Context) {
		userName := c.Param("user")

		db, err := readDb()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		for i, user := range db.Users {
			if user.Name == userName {
				c.JSON(http.StatusOK, gin.H{
					"exists":  true,
					"content": db.Users[i],
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"exists": false})
	})

	router.POST("/save/:user", func(c *gin.Context) {
		var UserData struct {
			Nodes []Node `json:"nodes"`
			Paths []Path `json:"paths"`
		}
		userName := c.Param("user")
		if err := c.Bind(&UserData); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			fmt.Println("Binding didn't work out", err.Error())
			return
		}

		db, err := readDb()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var userFound = false

		for i, user := range db.Users {
			if user.Name == userName {
				userFound = true
				db.Users[i].Nodes = UserData.Nodes
				db.Users[i].Paths = UserData.Paths
				fmt.Println("Found user, updating")
			}
		}

		if !userFound {
			db.Users = append(db.Users, User{
				Name:  userName,
				Nodes: UserData.Nodes,
				Paths: UserData.Paths,
			})
		}

		fmt.Println("Saving user now", db)

		saveDb(db)
		c.JSON(http.StatusOK, gin.H{
			"message": "data saved successfully",
		})
	})

	router.Run("localhost:5121")
}
