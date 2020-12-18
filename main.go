package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Staff is the representation of staff table
type Staff struct {
	ID           int
	FirstName    string `json:"first_name" validate:"required,gte=10"`
	LastName     string
	EmailID      string
	MobileNumber string `json:"password" validate:"required,gte=10"`
	Password     string
	Errors       map[string]string
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "picco123@$"
	dbName := "go_employee"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

// Login is gonna create new template
func Login(w http.ResponseWriter, r *http.Request) {

	tmpl.ExecuteTemplate(w, "Login", nil)
}

// Success is the representation of success page
func Success(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Success", nil)
}

// LoginProcess is defining the login functionality
func LoginProcess(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		emailID := r.FormValue("emailId")
		password := r.FormValue("password")

		// Validate form input
		if strings.Trim(emailID, " ") == "" || strings.Trim(password, " ") == "" {
			fmt.Println("Parameter's can't be empty")
			http.Redirect(w, r, "/", 301)
			return
		}

		checkUser, err := db.Query("SELECT id,password,first_name,last_name,email_id FROM Staff WHERE email_id=?", emailID)

		if err != nil {
			panic(err.Error())
		}
		staff := Staff{}
		for checkUser.Next() {
			var id int
			var password, firstName, lastName, emailID string
			err = checkUser.Scan(&id, &password, &firstName, &lastName, &emailID)
			if err != nil {
				panic(err.Error())
			}
			staff.ID = id
			staff.FirstName = firstName
			staff.LastName = lastName
			staff.EmailID = emailID
			staff.Password = password
		}

		fmt.Println(staff)

		errf := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(password))
		if errf != nil && errf == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
			fmt.Println(errf)
			http.Redirect(w, r, "/", 301)
		} else {
			fmt.Println("Success")

			fmt.Println(staff)
			tmpl.ExecuteTemplate(w, "Success", staff)
		}

	}
	defer db.Close()

}

// Register is the represntation of list page
func Register(w http.ResponseWriter, r *http.Request) {

	tmpl.ExecuteTemplate(w, "Register", nil)
}

// RegisterUser is the representation of user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
	if r.Method == "POST" {

		firstName := r.FormValue("firstname")
		lastName := r.FormValue("lastname")
		EmailID := r.FormValue("emailId")
		mobileNumber := r.FormValue("mobilenumber")
		fmt.Println(EmailID)
		password, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			tmpl.ExecuteTemplate(w, "Register", err)
		}

		dt := time.Now()
		//Format YYYY-MM-DD
		createdDate := dt.Format("2006-01-02 15:04:05")
		//confirmPassword := r.FormValue("confirmpassword")
		insForm, err := db.Prepare("INSERT INTO Staff(first_name, last_name,email_id,mobile_number,password,created_date) VALUES(?,?,?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(firstName, lastName, EmailID, mobileNumber, password, createdDate)
		log.Println("INSERT: First Name: " + firstName + " | Last Name: " + lastName)

	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func main() {
	log.Println("Server started on: http://localhost:8080")

	fs := http.FileServer(http.Dir("asset/"))
	http.Handle("/asset/", http.StripPrefix("/asset/", fs))

	http.HandleFunc("/", Login)
	http.HandleFunc("/loginsubmit", LoginProcess)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/registeruser", RegisterUser)
	http.HandleFunc("/success", Success)

	http.ListenAndServe(":8080", nil)
}
