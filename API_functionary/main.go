package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Functionary struct {
	Id         int    `json:"id"`
	Nome       string `json:"name"`
	Setor      string `json:"sector"`
	Email      string `json:"email"`
	Senha      string
	Permission int `json:"permission"`
}

const SecretKey = "blablabla"

func PasswordHash(s string) string {
	str := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", str)
}

var db *sql.DB

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("UpdateUser: Error to get cookie")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	token, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	result := db.QueryRow("SELECT AdmPermission FROM funcs WHERE Id = ?", claims.Issuer)

	var userDB Functionary

	errScan := result.Scan(&userDB.Permission)
	if errScan != nil {
		log.Println("User: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	if userDB.Permission != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Usuário não tem permissão"))
		return
	}

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

	userUpdated.Senha = PasswordHash(userUpdated.Senha)

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
		if userUpdated.Senha != "" {
			updatePassword, _ := db.Prepare("UPDATE funcs SET Password = ? WHERE Id = ?")
			updatePassword.Exec(userUpdated.Senha, id)
		}
		json.NewEncoder(w).Encode("Funcionário atualizado com sucesso!")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

func DeleteFunc(w http.ResponseWriter, r *http.Request) {
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("UpdateUser: Error to get cookie")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	token, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	result := db.QueryRow("SELECT AdmPermission FROM funcs WHERE Id = ?", claims.Issuer)

	var userDB Functionary

	errScan := result.Scan(&userDB.Permission)
	if errScan != nil {
		log.Println("User: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	if userDB.Permission != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Usuário não tem permissão"))
		return
	}

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
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("UpdateUser: Error to get cookie")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	token, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	result := db.QueryRow("SELECT AdmPermission FROM funcs WHERE Id = ?", claims.Issuer)

	var userDB Functionary

	errScan := result.Scan(&userDB.Permission)
	if errScan != nil {
		log.Println("User: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	if userDB.Permission != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Usuário não tem permissão"))
		return
	}

	body, errBody := io.ReadAll(r.Body)
	if errBody != nil {
		log.Println("InsertNewFunc: Error ao pegar body: ", errBody.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newUser Functionary

	errUnmarshal := json.Unmarshal(body, &newUser)
	if errUnmarshal != nil {
		log.Fatalln("Erro ao realizar Unmarshal do body para a struct")
	}

	newUser.Senha = PasswordHash(newUser.Senha)

	post, errPrep := db.Prepare("INSERT INTO funcs (Name, Sector, Email, Password, AdmPermission) VALUES (?, ?, ?, ?, ?)")
	if errPrep != nil {
		log.Fatalln("Erro prepare insert")
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, errInsert := post.Exec(newUser.Nome, newUser.Setor, newUser.Email, newUser.Senha, newUser.Permission)

	if errInsert != nil {
		log.Println("InsertNewFunc: Error ao realizar Insert: ", errInsert.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Funcionário cadastrado com sucesso!"))
	w.WriteHeader(http.StatusOK)
}

func ListOneFunc(w http.ResponseWriter, r *http.Request) {
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("UpdateUser: Error to get cookie")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	_, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	vars := mux.Vars(r)
	id, errConv := strconv.Atoi(vars["FuncId"])
	if errConv != nil {
		log.Println("ListOneFunc: Error ao realizar converção: ", errConv.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := db.QueryRow("SELECT Id, Name, Sector, Email, AdmPermission FROM funcs WHERE Id = ?", id)

	var funct Functionary

	errScan := result.Scan(&funct.Id, &funct.Nome, &funct.Setor, &funct.Email, &funct.Permission)
	if errScan != nil {
		log.Println("ListOneFunc: Error ao realizar scan: ", errScan.Error())
		w.Write([]byte("User not found"))
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
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("UpdateUser: Error to get cookie")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	_, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	result, errSelect := db.Query("SELECT Id, Name, Sector, Email, AdmPermission Email FROM funcs")
	if errSelect != nil {
		log.Println("ListAllFuncs: Error ao realizar select: ", errSelect.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var Funcs []Functionary = make([]Functionary, 0)

	for result.Next() {
		var funct Functionary
		errScan := result.Scan(&funct.Id, &funct.Nome, &funct.Setor, &funct.Email, &funct.Permission)
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

func Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalln("Error to get body of request")
	}

	var usrLogin Functionary

	json.Unmarshal(body, &usrLogin)

	var count int

	row, errSelect := db.Query("SELECT COUNT(*) FROM funcs WHERE Email = ?", usrLogin.Email)
	if errSelect != nil {
		log.Println("Login: Error ao buscar funcionário para login: ", errSelect.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for row.Next() {
		errScan := row.Scan(&count)
		if errScan != nil {
			log.Println("Login: Error ao realizar Scan: ", errScan.Error())
			return
		}
	}

	result := db.QueryRow("SELECT Id, Name, Sector, Email, Password FROM funcs WHERE Email = ?", usrLogin.Email)

	var FuncDB Functionary

	errScan := result.Scan(&FuncDB.Id, &FuncDB.Nome, &FuncDB.Setor, &FuncDB.Email, &FuncDB.Senha)
	if errScan != nil {
		log.Println("Login: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	if count == 1 {
		if FuncDB.Senha != PasswordHash(usrLogin.Senha) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Password invalid"))
		} else {
			claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
				Issuer:    strconv.Itoa(int(FuncDB.Id)),
				ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			})

			token, err := claims.SignedString([]byte(SecretKey))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Não foi possível fazer o login"))
				return
			}

			Cookie := http.Cookie{
				Name:     "jwt",
				Value:    token,
				Expires:  time.Now().Add(time.Hour * 24),
				HttpOnly: true,
			}
			http.SetCookie(w, &Cookie)

			w.Write([]byte("Sucess"))
		}
	}
}

func User(w http.ResponseWriter, r *http.Request) {
	cookie, errGetCookie := r.Cookie("jwt")
	if errGetCookie != nil {
		log.Println("User: Error to get cookie")
	}

	token, errGetToken := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if errGetToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User not loged"))
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	result := db.QueryRow("SELECT Id, Name, Sector, Email, Password FROM funcs WHERE Id = ?", claims.Issuer)

	var userDB Functionary

	errScan := result.Scan(&userDB.Id, &userDB.Nome, &userDB.Setor, &userDB.Email, &userDB.Senha)
	if errScan != nil {
		log.Println("User: Error ao realizar scan: ", errScan.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("User not found")
		return
	}

	json.NewEncoder(w).Encode(userDB)
	return
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	json.NewEncoder(w).Encode("message: sucess")
	return
}

func main() {
	ConfigDb()
	Router := mux.NewRouter().StrictSlash(true)
	Router.Use(jsonMiddleware)

	Router.HandleFunc("/funcs", ListAllFuncs).Methods("GET")
	Router.HandleFunc("/user", User).Methods("GET")
	Router.HandleFunc("/logout", LogOut).Methods("POST")

	Router.HandleFunc("/funcs/{FuncId}", ListOneFunc).Methods("GET")

	Router.HandleFunc("/funcs", InsertNewFunc).Methods("POST")

	Router.HandleFunc("/login", Login).Methods("POST")

	Router.HandleFunc("/funcs/{FuncId}", DeleteFunc).Methods("DELETE")

	Router.HandleFunc("/funcs/{FuncId}", UpdateUser).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", Router))
}
