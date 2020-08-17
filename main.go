package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sethvargo/go-password/password"
)

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

func main() {
	resource := "/signup"

	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}
		details := RegData{
			Email:             r.FormValue("email"),
			Password:          r.FormValue("password"),
			Portal:            r.FormValue("portal"),
			SendEmail:         r.FormValue("sendemail") == "true",
			PasswordGenerated: false,
			Success:           true,
		}
		if details.Password == "" {
			newpass, err := password.Generate(12, 2, 1, false, false)
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
				details.Problem = "User already exist"
				details.Unsuccess = true
				details.Success = false
			}
		} else {
			log.Fatal(err)
			details.Problem = err.Error()
			details.Unsuccess = true
			details.Success = false
		}

		tmpl.Execute(w, details)
	})
	http.ListenAndServe(":8099", nil)
}
