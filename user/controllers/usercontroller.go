package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/sharidas/go-petshotel/config"
	"github.com/sharidas/go-petshotel/user/models"

	"github.com/gorilla/mux"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UserUpdate struct {
	Email    string `json:"email"`
	UserData User   `json:"data"`
}

//Init : This function initialise the routes for the user
//	  API. Added routes here so that our server when starts
//	  has all the routes for user API.
func Init(r *mux.Router) {
	fmt.Println("Inside the init!!!")
	r.HandleFunc("/user/signup/{token}/", signupWithToken).Methods("GET")
	r.HandleFunc("/user/signup/", signup).Methods("POST")
	r.HandleFunc("/user/update/", updateUser).Methods("PUT")
}

func Login(email string, password string) error {
	return nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomStringUrlSafe(n int) (string, error) {
	fmt.Println("I am here!!!")
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

func sendmail() (string, error) {
	configData := &config.Configuration{}

	configData.ConfigParser()

	fromAddr := configData.MailSender
	smtpHost := configData.MailServer
	smtpPort := configData.MailPort
	passwd := configData.MailPasswd
	appPort := configData.AppPort

	name, _ := os.Hostname()
	fmt.Println("hostname = ", name)
	addrs, _ := net.LookupHost(name)

	var u url.URL
	for _, addr := range addrs {
		u = url.URL{Scheme: "http", Host: addr}
		fmt.Printf("type of u = %T\n", u)
		fmt.Println("url = ", u.String())
	}

	to := []string{
		string(configData.MailTo),
	}

	auth := smtp.PlainAuth("", fromAddr, passwd, smtpHost)
	url, urlErr := generateRandomStringUrlSafe(len(fromAddr) + 1)

	if urlErr != nil {
		fmt.Println("Error creating url: ", urlErr)
		return "", urlErr
	}
	fmt.Printf("The port is %v and type = %T\n", appPort, appPort)
	message := []byte("To signup, kindly click on the link " + u.String() + ":" + strconv.Itoa(appPort) + "/user/signup/" + url)
	//mailErr := smtp.SendMail(string(smtpHost)+":"+string(smtpPort), auth, fromAddr, to, message)

	mailErr := smtp.SendMail(smtpHost+":"+strconv.Itoa(smtpPort), auth, fromAddr, to, message)

	if mailErr != nil {
		fmt.Println("Mail send error: ", mailErr)
		return "", mailErr
	}

	//smtp.SendMail(smtpHost + ":" + smtpPort)

	return url, nil
}

func signup(w http.ResponseWriter, r *http.Request) {
	var data User

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//Send mail
	token, mailErr := sendmail()
	if mailErr != nil {
		fmt.Println("Error occured sending mail")
		return
	}

	username, password, email, role := data.Username, data.Password, data.Email, data.Role

	fmt.Printf("username = %v, password = %v, email = %v, role = %v\n", username, password, email, role)

	// hash the passwd
	hashPasswd, errPasswd := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errPasswd != nil {
		fmt.Println("Password error: ", errPasswd)
		return
	}

	//Populate the UserModel struct
	userModel := &models.UserModel{Username: username, Password: string(hashPasswd), Email: email, Role: role, Token: token}
	_, err1 := userModel.CreateUser()

	if err1 != nil {
		fmt.Println("Error: ", err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Something went wrong\n")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "All is ok at homehandler\n")
}

func signupWithToken(w http.ResponseWriter, r *http.Request) {
	var data User

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := data.Email

	vars := mux.Vars(r)
	tokenRequested := vars["token"]
	fmt.Println("The token obtained = ", tokenRequested)

	userModel := &models.UserModel{Email: email, Token: tokenRequested}
	status, errVerify := userModel.VerifyToken()

	if errVerify != nil {
		fmt.Println("Something wrong happened")
	}
	fmt.Println("status = ", status)
	if errVerify != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Wrong token provided\n")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Correct handler for token\n")
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	/*var data User

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}*/

	//reqData := make(map[string]interface{})
	//var reqData UserUpdate
	//reqData := map[string]interface{}{}
	//var reqData interface{}
	//reqData := &UserUpdate{}
	var reqData UserUpdate

	//err := json.NewDecoder(r.Body).Decode(&reqData)
	//decoder := json.NewDecoder(r.Body)
	//decoder.DisallowUnknownFields()
	decodeErr := json.NewDecoder(r.Body).Decode(&reqData)

	fmt.Printf("data got from the user = %v\n", reqData)

	username, passwd, ok := r.BasicAuth()
	fmt.Printf("username = %v, password = %v, ok = %v\n", username, passwd, ok)

	if decodeErr != nil {
		fmt.Println("Error: ", decodeErr)
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	if username == "" || passwd == "" {
		fmt.Printf("Login credentials required\n")
		return
	}

	//Create model and call the login

	usermodel := &models.UserModel{Email: username, Password: passwd}
	loginErr := usermodel.LoginUser()
	if loginErr != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Wrong credentials provided\n")
		return
	}
}
