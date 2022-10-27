package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID    int
	Nome  string
	Email string
	Idade int
}

var db *sql.DB

var Users []User

func CadastrarUser(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var NewUser User
	json.Unmarshal(body, &NewUser)
	NewUser.ID = len(Users) + 1
	Users = append(Users, NewUser)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(NewUser)
}

func updateUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["userId"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var userUpdated User

	errJson := json.Unmarshal(body, &userUpdated)
	if errJson != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	indiceUser := -1

	for i, usr := range Users {
		if usr.ID == id {
			indiceUser = i
			break
		}
	}

	if indiceUser < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if userUpdated.Email != "" {
		Users[indiceUser].Email = userUpdated.Email
	}
	if userUpdated.Idade != 0 {
		Users[indiceUser].Idade = userUpdated.Idade
	}
	if userUpdated.Nome != "" {
		Users[indiceUser].Nome = userUpdated.Nome
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(Users[indiceUser])
}

func GetOneUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["userId"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	result := db.QueryRow("SELECT id, name, email, age FROM users WHERE id = ?", id)

	var usr User

	errScan := result.Scan(&usr.ID, &usr.Nome, &usr.Email, &usr.Idade)
	if errScan != nil {
		log.Println("GetOneUser: Scan: ", errScan.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errEnconde := json.NewEncoder(w).Encode(usr)
	if errEnconde != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("GetOneUser: Encode: ", errEnconde.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["userId"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, errSelect := db.Query("DELETE FROM users WHERE id = ?", id)
	if errSelect != nil {
		log.Println("Delete: ", errSelect.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errEnconde := json.NewEncoder(w).Encode("User deletado com sucesso")
	if errEnconde != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Delete: Encode: ", errEnconde.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func ListarUser(w http.ResponseWriter, r *http.Request) {

	result, errSelect := db.Query("SELECT id, name, email, age FROM users")
	if errSelect != nil {
		log.Println("ListarUser: ", errSelect.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var usuarios []User = make([]User, 0)

	for result.Next() {
		var user User
		errScan := result.Scan(&user.ID, &user.Nome, &user.Email, &user.Idade)
		if errScan != nil {
			log.Println("ListarUser: Scan: ", errScan.Error())
			continue
		}

		usuarios = append(usuarios, user)
	}

	errCloseResult := result.Close()
	if errCloseResult != nil {
		log.Println("ListarUser: CloseResult: ", errCloseResult.Error())
		return
	}

	json.NewEncoder(w).Encode(usuarios)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func configDb() {
	var errAbertura error
	db, errAbertura = sql.Open("mysql", "root:94647177_Mc@tcp(localhost:3306)/crud")
	if errAbertura != nil {
		log.Fatalln("Erro ao conectar com o banco: ", errAbertura.Error())
		return
	}

	errConexDB := db.Ping()
	if errConexDB != nil {
		log.Fatalln(errConexDB.Error())
	}
}

func main() {
	configDb()
	roteador := mux.NewRouter().StrictSlash(true)
	roteador.Use(jsonMiddleware)
	roteador.HandleFunc("/user", ListarUser).Methods("GET")
	roteador.HandleFunc("/user/{userId}", GetOneUser).Methods("GET")
	roteador.HandleFunc("/user/{userId}", deleteUser).Methods("DELETE")
	roteador.HandleFunc("/user/{userId}", updateUser).Methods("PUT")
	roteador.HandleFunc("/user", CadastrarUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", roteador))
}
