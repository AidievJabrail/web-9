package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"io"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)
const (
	host     = "localhost"
	port     = 5432
	user     = "jabrail"
	password = "07012006"
	dbname   = "querydb"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

type myStruct struct {
	Count int `json:"count"`
}


func (h Handlers) handler(c echo.Context) error{
	switch c.Request().Method {
	case http.MethodGet:
		count, err := h.dbProvider.GetCount()
		if err != nil{
			return c.String(http.StatusBadRequest, "Не удалось получить count") 
        	
		}
		return c.String(http.StatusOK, count)
	case http.MethodPost:
		
		var tmp myStruct
		c.Request().ParseForm()
		data, _ := io.ReadAll(c.Request().Body)
		if err := json.Unmarshal(data, &tmp); err != nil {
			return c.String(http.StatusBadRequest, "Не получилось открыть данные")
		}
		err := h.dbProvider.IncrementCount(tmp.Count)
		if err != nil{
			return c.String(http.StatusBadRequest, "Не удалось увеличить count")
		}
		return c.String(http.StatusOK, "Успешно!")
	}
	return c.String(http.StatusOK, "Успешно!")
}

func (h Handlers) SetCount(c echo.Context) error{
	var tmp myStruct
	c.Request().ParseForm()
	data, _ := io.ReadAll(c.Request().Body)
	if err := json.Unmarshal(data, &tmp); err != nil {
		return c.String(http.StatusBadRequest, "Не получилось открыть данные")
	}
	err := h.dbProvider.SetCountSQL(tmp.Count)
	if err != nil{
		return c.String(http.StatusBadRequest, "Не удалось увеличить count")
	}
	return c.String(http.StatusOK, "Успешно!")
}

func (dp DatabaseProvider) IncrementCount(num int) error{
	_, err := dp.db.Exec("UPDATE countdb SET count = count + $1", num)
	if err != nil {
		return err
	}
	return nil
}

func (dp DatabaseProvider) SetCountSQL(num int) error{
	_, err := dp.db.Exec("UPDATE countdb SET count = $1", num)
	if err != nil {
		return err
	}
	return nil
}

func (dp DatabaseProvider) GetCount() (string, error){
	var resp string
	row := dp.db.QueryRow("SELECT count FROM countdb")
	err := row.Scan(&resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}

	h := Handlers{dbProvider: dp}
	e := echo.New()

	e.Logger.SetLevel(2)

    e.GET("/count", h.handler)
	e.POST("/count", h.handler)
	e.POST("/count/set", h.SetCount)
	if err := e.Start(*address); err != nil{
		log.Fatal(err)
	}
}