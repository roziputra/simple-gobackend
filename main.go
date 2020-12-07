package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Category struct {
	Kode string `json:"kode"`
	Name string `json:"name"`
}

type Book struct {
	Kode     string `json:"kode"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Format   string `json:"format"`
	Price    string `json:"price"`
	Date     string `json:"date"`
}

type ResponseCategories struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []Category
}

type ResponseBooks struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Total   int    `json:"Total"`
	Page    int    `json:"Page"`
	Data    []Book
}

func returnAllCategories(w http.ResponseWriter, r *http.Request) {

	db := dbConn()

	var category Category
	var Categories []Category
	var response ResponseCategories

	rows, err := db.Query("select distinct kode,name from categories")
	if err != nil {
		log.Print(err)
	}

	for rows.Next() {
		if err := rows.Scan(&category.Kode, &category.Name); err != nil {
			log.Fatal(err.Error())
		} else {
			Categories = append(Categories, category)
		}
	}

	response.Status = 1
	response.Message = "Success"
	response.Data = Categories

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)
	fmt.Println("Endpoint Hit: returnAllCategories")
}

func returnBookbyCategory(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	kodecategory := vars["kode"]

	db := dbConn()

	var book Book
	var Books []Book
	var response ResponseBooks

	rows, err := db.Query("Select b.kode, b.title, c.name category, b.format, b.price, b.date from books b left join categories c on b.category=c.id where c.kode=?", kodecategory)

	if err != nil {
		log.Print(err)
	}

	for rows.Next() {
		if err := rows.Scan(&book.Kode, &book.Title, &book.Category, &book.Format, &book.Price, &book.Date); err != nil {
			log.Fatal(err.Error())
		} else {
			Books = append(Books, book)
		}
	}

	response.Status = 1
	response.Message = "Success"
	response.Data = Books

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)
	fmt.Println("Endpoint Hit: returnBookbyCategory")
}

func returnBookbyKode(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	kodebook := vars["kode"]

	db := dbConn()

	var book Book
	var Books []Book
	var response ResponseBooks

	sqlQuery := " Select b.kode, b.title, GROUP_CONCAT(c.name SEPARATOR ', ') category, b.format, b.price, date(b.date) from books b " +
		" left join categories c on b.category=c.id where b.kode=? " +
		" group by b.kode, b.title, b.format, b.price, date(b.date) "

	rows, err := db.Query(sqlQuery, kodebook)

	if err != nil {
		log.Print(err)
	}

	for rows.Next() {
		if err := rows.Scan(&book.Kode, &book.Title, &book.Category, &book.Format, &book.Price, &book.Date); err != nil {
			log.Fatal(err.Error())
		} else {
			Books = append(Books, book)
		}
	}

	response.Status = 1
	response.Message = "Success"
	response.Data = Books

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)
	fmt.Println("Endpoint Hit: returnBookbyKode")
}

func returnBookbyFilter(w http.ResponseWriter, r *http.Request) {

	titlePost := ""
	categoryPost := 0
	minPricePost := -1
	maxPricePost := -1
	pagePost := 1

	if err := r.ParseForm(); err != nil {
		titlePost = ""
		categoryPost = 0
		minPricePost = -1
		maxPricePost = -1
		pagePost = 1
	} else {
		titlePost = r.FormValue("title")

		categoryPost, err = strconv.Atoi(r.FormValue("category"))
		if err != nil {
			categoryPost = 0
		}

		strMinPricePost := r.FormValue("price[min]")
		if strMinPricePost == "" {
			minPricePost = -1
		} else {
			minPricePost, err = strconv.Atoi(strMinPricePost)
			if err != nil {
				minPricePost = -1
			}
		}

		strMaxPricePost := r.FormValue("price[max]")
		if strMaxPricePost == "" {
			maxPricePost = -1
		} else {
			maxPricePost, err = strconv.Atoi(strMaxPricePost)
			if err != nil {
				maxPricePost = -1
			}
		}

		pagePost, err = strconv.Atoi(r.FormValue("page"))
		if err != nil {
			pagePost = 1
		}
	}

	db := dbConn()

	var book Book
	var Books []Book
	var response ResponseBooks

	totalcount := 0

	limit := 100
	offset := (pagePost - 1) * limit

	filterquery := "Select b.kode, b.title, c.name category, b.format, b.price, b.date from books b left join categories c on b.category=c.id where 1 "
	totalquery := "select count(b.id) total from books b left join categories c on b.category=c.id where 1 "

	if titlePost != "" {
		filterquery = filterquery + " and b.title like '%" + titlePost + "%' "
		totalquery = totalquery + " and b.title like '%" + titlePost + "%' "
	}
	if categoryPost != 0 {
		filterquery = filterquery + " and c.kode = '" + strconv.Itoa(categoryPost) + "' "
		totalquery = totalquery + " and c.kode = '" + strconv.Itoa(categoryPost) + "' "
	}
	if minPricePost >= 0 {
		filterquery = filterquery + " and CAST(b.price AS DECIMAL(10,2)) >= '" + strconv.Itoa(minPricePost) + "' "
		totalquery = totalquery + " and CAST(b.price AS DECIMAL(10,2)) >= '" + strconv.Itoa(minPricePost) + "' "
	}
	if maxPricePost >= 0 {
		filterquery = filterquery + " and CAST(b.price AS DECIMAL(10,2)) <= '" + strconv.Itoa(maxPricePost) + "' "
		totalquery = totalquery + " and CAST(b.price AS DECIMAL(10,2)) <= '" + strconv.Itoa(maxPricePost) + "' "
	}

	pagequery := filterquery + " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)

	total, err := db.Query(totalquery)
	if err != nil {
		log.Print(err)
	}
	for total.Next() {
		if err := total.Scan(&totalcount); err != nil {
			log.Fatal(err.Error())
		}
	}

	rows, err := db.Query(pagequery)
	if err != nil {
		log.Print(err)
	}
	for rows.Next() {
		if err := rows.Scan(&book.Kode, &book.Title, &book.Category, &book.Format, &book.Price, &book.Date); err != nil {
			log.Fatal(err.Error())
		} else {
			Books = append(Books, book)
		}
	}

	response.Status = 1
	response.Message = "Success"
	response.Total = totalcount
	response.Page = pagePost
	response.Data = Books

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)
	fmt.Println("Endpoint Hit: returnBookbyFilter")
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "rozi"
	dbName := "scrap"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Golang Backend!")
}

func handleRequests() {

	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)

	var api = myRouter.PathPrefix("/api").Subrouter()

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	var api1 = api.PathPrefix("/v1").Subrouter()

	api1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	// replace http.HandleFunc with api1.HandleFunc
	api1.HandleFunc("/", homePage)

	api1.HandleFunc("/books/categories", returnAllCategories)

	api1.HandleFunc("/books/category/{kode}", returnBookbyCategory)

	api1.HandleFunc("/book/detail/{kode}", returnBookbyKode)

	api1.HandleFunc("/books/filter", returnBookbyFilter)

	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {

	fmt.Println("listen to port 8080")
	handleRequests()
}
