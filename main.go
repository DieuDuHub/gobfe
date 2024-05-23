package main

import (
	"encoding/json"
	"fmt"

	"io"

	"github.com/gin-gonic/gin"

	"log"

	"net/http"
	os "os"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	response, err := http.Get("http://localhost/referential/dcx")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseData))

	c.String(http.StatusOK, string(responseData))
}

func getRedirect(c *gin.Context, destination string) {
	response, err := http.Get(destination)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseData))

	c.String(http.StatusOK, string(responseData))
}

type Redirect struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Configuration struct {
	Redirect []Redirect `json:"redirect"`
}

func loadConfig() Configuration {
	response, err := http.Get("http://localhost/config/rustmo/default/main/gobfe.json")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close()
	/*
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	*/
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	//file, _ := os.Open("conf.json")
	//	defer file.Close()
	//configuration := Configuration{}

	var c Configuration
	err = json.Unmarshal(body, &c)
	//configuration := Configuration{}
	//	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	/*for _, i := range c {
		fmt.Println(i.referential)
		fmt.Println(i.url)

	}*/
	fmt.Printf("%+v\n", c)

	return c
}

func main() {
	fmt.Println("Hello, World!")

	config := loadConfig()

	router := gin.Default()

	//var keycloakurl string = "http://localhost/realms/ecommerce/protocol/openid-connect/certs"

	keycloakurl := os.Getenv("jwt_token_rsa_url")

	var rsakey, err = GetJWKSet(keycloakurl)
	if err != nil {
		fmt.Println("Error lodaing RSA Public Key from env var : jwt_token_rsa_url ")
		fmt.Println(err)
		os.Exit(1)
	}
	router.Use(JWTMiddleware(rsakey))

	router.GET("/albums", getAlbums)

	for _, i := range config.Redirect {
		fmt.Println(i.Url)
		router.GET(i.Name, func(c *gin.Context) {
			getRedirect(c, i.Url)
		})
	}

	router.Run("localhost:8080")

}
