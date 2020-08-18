package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strings"

	pswd "github.com/sethvargo/go-password/password"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	auth "github.com/korylprince/go-ad-auth"
)

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// RegData данные для регистрации
type RegData struct {
	Email             string
	Password          string
	Portal            string
	SendEmail         bool
	PasswordGenerated bool
	Success           bool
	Problem           string
	Unsuccess         bool
	ADuser            string
	ADPassword        string
	IP                string
}

type RegResult struct {
	Result   int
	Response struct {
		Result    int
		App       string
		Exist     bool
		UID       int
		DateTime  string
		Email     string
		Activated bool
	}
}

// isEmailValid
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	if !emailRegex.MatchString(e) {
		return false
	}
	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

func main() {

	config := &auth.Config{
		Server:   "10.55.0.98",
		Port:     389,
		BaseDN:   "DC=cloud,DC=local",
		Security: auth.SecurityInsecureStartTLS,
	}

	resource := "/signup"

	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			details := RegData{
				Portal:    "DataFort",
				SendEmail: false,
			}
			tmpl.Execute(w, details)
			return
		}
		details := RegData{
			Email:             r.FormValue("email"),
			Password:          r.FormValue("password"),
			Portal:            r.FormValue("portal"),
			SendEmail:         r.FormValue("sendemail") == "true",
			PasswordGenerated: false,
			Success:           true,
			ADuser:            r.FormValue("aduser"),
			ADPassword:        r.FormValue("adpassword"),
			IP:                r.RemoteAddr,
		}

		status, err := auth.Authenticate(config, details.ADuser, details.ADPassword)

		if err != nil {
			log.Fatal(err)
		}
		if !isEmailValid(details.Email) {
			details.Problem = "Несуществующий email"
			details.Unsuccess = true
			details.Success = false
			log.Printf("User %v from host %v try registr bad email", details.ADuser, details.IP)
		} else if !status {
			details.Problem = "Мы вас не узнали"
			details.Unsuccess = true
			details.Success = false
			log.Printf("Auth fails for user %v from host %v", details.ADuser, details.IP)
		} else {
			if details.Password == "" {
				newpass, err := pswd.Generate(12, 2, 1, false, false)
				if err != nil {
					log.Fatal(err)
				}
				details.Password = newpass
				details.PasswordGenerated = true
			}
			data := url.Values{}
			data.Set("appid", "signup")
			data.Set("email", details.Email)
			data.Set("password", details.Password)
			data.Set("SendEmail", strconv.FormatBool(details.SendEmail))
			apiURL := "https://reg.paasinfra.datafort.ru"

			switch details.Portal {
			case "DataFort":
				apiURL = "https://reg.paasinfra.datafort.ru"
			case "Beeline":
				apiURL = "https://reg.paas.beelinecloud.ru/"
			case "SysSoft":
				apiURL = "https://reg.paas.syssoft.ru"
			}
			u, _ := url.ParseRequestURI(apiURL)
			u.Path = resource
			urlStr := u.String()
			client := &http.Client{}
			buffer := new(bytes.Buffer)
			buffer.WriteString(data.Encode())
			r2, _ := http.NewRequest(http.MethodPost, urlStr, buffer) // URL-encoded payload
			r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r2.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))

			resp, err := client.Do(r2)
			if err == nil {
				responseData, _ := ioutil.ReadAll(resp.Body)
				var RegResult RegResult
				json.Unmarshal(responseData, &RegResult)
				if RegResult.Response.Exist {
					details.Problem = "Такой пользователь уже сушествует"
					details.Unsuccess = true
					details.Success = false
					log.Printf("User %v from host %v try to create allready existing account (%v)", details.ADuser, details.IP, details.IP)
				} else {
					if details.PasswordGenerated {
						log.Printf("User %v from host %v tcreate account (%v) with generated password", details.ADuser, details.IP, details.IP)
					} else {
						log.Printf("User %v from host %v tcreate account (%v)", details.ADuser, details.IP, details.IP)
					}

				}
			} else {
				log.Fatal(err)
				details.Problem = err.Error()
				details.Unsuccess = true
				details.Success = false
			}
		}
		tmpl.Execute(w, details)
	})
	http.ListenAndServe(":8099", nil)
}
