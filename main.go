package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/twocaptcha"
)

type config struct {
	Captcha  string `json:"captcha_api_key"`
	Catchall string `json:"catchall"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type usernameCheck struct {
	Taken bool `json:"success"`
}

type register struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Recaptcha string `json:"recaptcha"`
	Username  string `json:"username"`
}

var availableNames = make([]string, 0)
var claimedNames = make([]string, 0)
var fig config
var twoCap *twocaptcha.TwoCaptchaClient
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	file, err := os.Open("wordlist.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	usernames := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		usernames = append(usernames, scanner.Text())
	}

	content, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	json.Unmarshal(content, &fig)

	var wg sync.WaitGroup

	for _, name := range usernames {
		if len(name) < 3 { // Invalid username length
			continue
		}

		wg.Add(1)
		go checkIfAvailable(name, &wg)
	}

	wg.Wait()
	twoCap = twocaptcha.New(fig.Captcha)

	var wg2 sync.WaitGroup

	for _, n := range availableNames {
		wg2.Add(1)
		go registerAccount(n, &wg2)
	}

	wg2.Wait()

	for _, n := range claimedNames {
		file, err := os.OpenFile("claimed.txt", os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		if _, err := file.WriteString(n); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Claimed " + strconv.Itoa(len(claimedNames)) + " account(s) and saved to file.")

}

func checkIfAvailable(n string, wg *sync.WaitGroup) {
	defer wg.Done()
	n = strings.ToLower(n) // Hiven appears to only save names as lowercase, and the /users endpoint is case sensitive

	resp, err := http.Get("https://api.hiven.io/v1/users/" + n)
	if err != nil {
		fmt.Println("Error checking username - ", n, err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var success usernameCheck
	err = json.Unmarshal(body, &success)
	if err != nil {
		fmt.Println("Error checking username - ", n, err)
		return
	}

	if success.Taken == true {
		return
	}

	availableNames = append(availableNames, n)
	fmt.Println(n + " is available")
}

func registerAccount(n string, wg *sync.WaitGroup) {
	fmt.Println("Attempting to register " + n)
	defer wg.Done()

	data := register{
		Email:    n + "@" + fig.Catchall,
		Name:     fig.Name,
		Password: fig.Password,
		Username: n,
	}

	cap, err := twoCap.SolveRecaptchaV2("https://canary.hiven.io/auth/register", "6Leu1rIUAAAAAOP9NG2ewnK1F7bbC51PFPqMvZXQ")
	if err != nil {
		panic(err)
	}

	data.Recaptcha = cap
	by, _ := json.Marshal(data)
	resp, err := http.Post("https://api.hiven.io/v1/auth/register", "application/json", bytes.NewBuffer(by))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)

	type successCheck struct {
		Success bool `json:"success"`
	}
	var su successCheck
	json.Unmarshal(b, &su)

	if su.Success == false {
		fmt.Println("Failed to claim username " + n)
		return
	}

	fmt.Println("Claimed username " + n)
	claimedNames = append(claimedNames, n)
}
