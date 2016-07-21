package main

import (

	//"database/sql"
	//"log"

	"time"

	"github.com/gin-gonic/gin"
	//"gopkg.in/gorp.v1"
	//_ "github.com/mattn/go-sqlite3"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/satori/go.uuid"
)

type RawData struct {
	Value json.Number `json: "Value"`
	Time  time.Time   `json: "Time"`
}

type Connection struct {
	Mc *memcache.Client
}

type Node struct {
	Id      string `db:"article_id"`
	Created time.Time
	Title   string
	Data    RawData `json: "Data"`
}

func createNode(title string, value json.Number, uuid string, i int) Node {
	return Node{Id: uuid, Created: time.Now(), Title: title, Data: RawData{Time: time.Now().Add(time.Duration(i) * time.Second), Value: value}}
}

func createNodes(title string, value json.Number) ([]Node, error) {

	nodes := make([]Node, 2)

	id := uuid.NewV4().String()
	//nodes = append(nodes, createNode(title, value, id, 20))
	nodes[0] = createNode(title, value, id, 10)
	//fmt.Println(nodes[0].Id)

	return nodes, nil
}

func viewHandler(c *gin.Context) {

	title := c.Request.URL.Path[len("/view/"):]

	var Mc *memcache.Client = newMemcache()

	item, err := Mc.Get(title)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var node []Node

	err = json.Unmarshal(item.Value, &node)

	fmt.Println(node)
	if len(node) == 0 {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Didn't append right"))
		return
	}
	content := gin.H{
		"Id":      node[0].Id,
		"Created": node[0].Created,
		"Title":   node[0].Title,
		"Data":    node[0].Data,
	}

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	//fmt.Println(node.Id)
	//fmt.Fprintf(c.Writer, "<h1>Hello world</h1>")
	c.IndentedJSON(200, content)
}

func formHandler(c *gin.Context) {
	t, err := template.ParseFiles("form.html")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	t.Execute(c.Writer, nil)

}

func newMemcache() *memcache.Client {
	return memcache.New("localhost:51030")
}

func saveHandler(c *gin.Context) {

	t := c.Request.FormValue("title")

	v := c.Request.FormValue("value")

	var Mc *memcache.Client = newMemcache()

	//vf, err := strconv.ParseFloat(v, 64)

	/*if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}*/
	vf := json.Number(v)

	_, err := Mc.Get(t)

	if err != nil {
		//id := uuid.NewV4().String()

		//fmt.Println(id)
		//node := Node{Id: id, Created: time.Now(), Title: t, Data: RawData{Value: vf, Time: time.Now()}}
		node, _ := createNodes(t, vf)

		b, err := json.Marshal(node)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		item := memcache.Item{Key: t, Value: b}

		err = Mc.Add(&item)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)

			return
		} else {
			c.Redirect(http.StatusFound, "/view/"+t)
		}

	} else {
		c.Redirect(http.StatusFound, "/view/"+t)
	}

}

/*func initDb() *gorp.DbMap {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	checkErr(err, "sql.Open failed")
	dbmap := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.SqliteDialect{},
	}

	dbmap.AddTableWithName(Node{}, "nodes").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}
func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}*/

func main() {

	router := gin.Default()
	router.LoadHTMLFiles("index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	router.GET("/view/:title", viewHandler)
	router.GET("/form/", formHandler)
	router.POST("/save/", saveHandler)
	router.Run(":8080")
}
