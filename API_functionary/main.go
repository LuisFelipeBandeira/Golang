package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

type Functionary struct {
	Id    int
	Nome  string
	Setor string
	Email string
}

var db *sql.DB

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, errConv := strconv.Atoi(vars["FuncId"])
	if errConv != nil {
		log.Println("UpdateUser: Error ao realizar converção: ", errConv.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, errBody := io.ReadAll(r.Body)
	if errBody != nil {
		log.Println("UpdateUser: Error ao pegar body: ", errBody.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var userUpdated Functionary

	json.Unmarshal(body, &userUpdated)

	row, errSelectContent := db.Query("SELECT COUNT(*) FROM funcs WHERE Id = ?", id)
	if errSelectContent != nil {
		log.Println("UpdateUser: Error ao buscar funcionário para ser atualizado: ", errSelectContent.Error())
		w.WriteHeader(http.StatusNotFound)
	}

	var count int

	for row.Next() {
		errScan := row.Scan(&count)
		if errScan != nil {
			log.Println("UpdateUser: Error ao realizar Scan: ", errScan.Error())
			return
		}
	}

	if count == 1 {
		if userUpdated.Nome != "" {
			updateName, _ := db.Prepare("UPDATE funcs SET Name = ? WHERE Id = ?")
			updateName.Exec(userUpdated.Nome, id)
		}
		if userUpdated.Setor != "" {
			updateSetor, _ := db.Prepare("UPDATE funcs SET Sector = ? WHERE Id = ?")
			updateSetor.Exec(userUpdated.Setor, id)
		}
		if userUpdated.Email != "" {
			updateEmail, _ := db.Prepare("UPDATE funcs SET Email = ? WHERE Id = ?")
			updateEmail.Exec(userUpdated.Email, id)
		}
		json.NewEncoder(w).Encode("Funcionário atualizado com sucesso!")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		log.Println("Está parando aqui")
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

func DeleteFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, errConv := strconv.Atoi(vars["FuncId"])
	if errConv != nil {
		log.Println("DeleteFunc: Error ao realizar converção: ", errConv.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete, _ := db.Prepare("DELETE FROM funcs WHERE Id = ?")
	_, errDelete := delete.Exec(id)
	if errDelete != nil {
		log.Println("DeleteFunc: Error ao realizar DELETE: ", errDelete.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Funcionário deletado com sucesso!")
	w.WriteHeader(http.StatusOK)
}

func InsertNewFunc(w http.ResponseWriter, r *http.Request) {
	body, errBody := io.ReadAll(r.Body)
	if errBody != nil {
		log.Println("InsertNewFunc: Error ao pegar body: ", errBody.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newUser Functionary

	json.Unmarshal(body, &newUser)

	post, _ := db.Prepare("INSERT INTO funcs (Name, Sector, Email) VALUES (?, ?, ?)")

	_, errInsert := post.Exec(newUser.Nome, newUser.Setor, newUser.Email)

	if errInsert != nil {
		log.Println("InsertNewFunc: Error ao realizar Insert: ", errInsert.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Funcionário cadastrado com sucesso!")
	w.WriteHeader(http.StatusOK)
}

func ListOneFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, errConv := strconv.Atoi(vars["FuncId"])
	if errConv != nil {
		log.Println("ListOneFunc: Error ao realizar converção: ", errConv.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := db.QueryRow("SELECT Id, Name, Sector, Email FROM funcs WHERE Id = ?", id)

	var funct Functionary

	errScan := result.Scan(&funct.Id, &funct.Nome, &funct.Setor, &funct.Email)
	if errScan != nil {
		log.Println("ListOneFunc: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	errEnconde := json.NewEncoder(w).Encode(funct)
	if errEnconde != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("ListOneFunc: Encode: ", errEnconde.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ListAllFuncs(w http.ResponseWriter, r *http.Request) {

	result, errSelect := db.Query("SELECT Id, Name, Sector, Email FROM funcs")
	if errSelect != nil {
		log.Println("ListAllFuncs: Error ao realizar select: ", errSelect.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var Funcs []Functionary = make([]Functionary, 0)

	for result.Next() {
		var funct Functionary
		errScan := result.Scan(&funct.Id, &funct.Nome, &funct.Setor, &funct.Email)
		if errScan != nil {
			log.Println("ListAllFuncs: Error ao realizar scan: ", errScan.Error())
			continue
		}

		Funcs = append(Funcs, funct)
	}

	errClose := result.Close()
	if errSelect != nil {
		log.Println("ListAllFuncs: Error ao realizar close: ", errClose.Error())
		return
	}

	errEnconde := json.NewEncoder(w).Encode(Funcs)
	if errEnconde != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("ListAllFuncs: Encode: ", errEnconde.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ConfigDb() {
	var errConnect error
	db, errConnect = sql.Open("mysql", "root:94647177_Mc@tcp(localhost:3306)/functionarys")
	if errConnect != nil {
		log.Fatalln("Erro ao conectar com o banco: ", errConnect.Error())
		return
	}

	errPing := db.Ping()
	if errPing != nil {
		log.Fatalln("Erro ao pingar DB", errPing.Error())
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	ConfigDb()
	Router := mux.NewRouter().StrictSlash(true)
	Router.Use(jsonMiddleware)
	Router.HandleFunc("/funcs", ListAllFuncs).Methods("GET")
	Router.HandleFunc("/funcs/{FuncId}", ListOneFunc).Methods("GET")
	Router.HandleFunc("/funcs", InsertNewFunc).Methods("POST")
	Router.HandleFunc("/funcs/{FuncId}", DeleteFunc).Methods("DELETE")
	Router.HandleFunc("/funcs/{FuncId}", UpdateUser).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", Router))
}
