package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/igorek306/onesecmail"
)

func checkErr(err error) bool {
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	return err != nil
}

func confHeader(req *http.Request) {
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "pl;q=0.7")
	req.Header.Add("content-type", "application/json;charset=UTF-8")
	req.Header.Add("origin", "https://account.atlasvpn.com")
	req.Header.Add("referer", "https://account.atlasvpn.com/")
	req.Header.Add("sec-ch-ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Brave\";v=\"120\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-site")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Add("x-client-id", "Web")
}

type TokenRes struct {
	Token string `json:"token"`
}

type AtlasAccount struct {
	MailAddr          string
	Auth              string
	LastEmailReadDate string
}

type AtlasApiUser struct {
	Uuid string `json:"uuid"`
}

func (a *AtlasAccount) fetchUuid(client *http.Client) (string, error) {
	url := "https://user.atlasvpn.com/v3/user"
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	confHeader(req)
	req.Header.Add("Authorization", a.Auth)
	if checkErr(err) {
		return "", err
	}
	res, err := client.Do(req)
	if checkErr(err) {
		return "", err
	}
	defer res.Body.Close()
	d, err := bodyToBytes(res.Header.Get("Content-Encoding"), &res.Body)
	if checkErr(err) {
		return "", err
	}
	var user AtlasApiUser
	err = json.Unmarshal(d, &user)
	if checkErr(err) {
		return "", err
	}

	return user.Uuid, nil
}

func bodyToBytes(contentEncoding string, res *io.ReadCloser) ([]byte, error) {
	var d []byte
	var err error
	if contentEncoding == "gzip" {
		reader, err := gzip.NewReader(*res)
		if checkErr(err) {
			return d, err
		}
		d, err = io.ReadAll(reader)
		return d, err
	} else {
		d, err = io.ReadAll(*res)
		return d, err

	}
}

func codeToToken(code string, client *http.Client) (TokenRes, error) {
	var token TokenRes
	url := "https://user.atlasvpn.com/v1/tokens/" + code
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if checkErr(err) {
		fmt.Println("Error creating request for join endpoint")
		return token, err
	}
	confHeader(req)

	res, err := client.Do(req)
	if checkErr(err) {
		fmt.Println("Error sending HTTP request to tokens endpoint")
		return token, err
	}
	d, err := bodyToBytes(res.Header.Get("Content-Encoding"), &res.Body)

	if checkErr(err) {
		fmt.Println("Error converting response body to bytes")
		return token, err
	}
	err = json.Unmarshal(d, &token)
	if checkErr(err) {
		fmt.Println("Error decoding tokens json")
		return token, err
	}
	res.Body.Close()
	return token, nil
}

func generateAtlasVPNAccount(referral string, mailClient *onesecmail.Client, client *http.Client, tunnel chan bool) (AtlasAccount, error) {
	aa := AtlasAccount{}

	defer func() {

		if tunnel != nil {
			tunnel <- true
		}
	}()
	addresses, err := mailClient.GenerateRandomEmailAddresses(1)
	if checkErr(err) {
		fmt.Println("Error getting all active domains")
		return aa, err
	}
	address := addresses[0]
	defer func() { checkErr(mailClient.ClearMailbox(address)) }()
	url := "https://user.atlasvpn.com/v1/request/join"
	method := "POST"

	payload := strings.NewReader(`{
   "email": "` + address + `",
   "marketing_consent": true,
   "referrer_uuid": "` + referral + `",
   "referral_offer": "initial"
  }`)
	req, err := http.NewRequest(method, url, payload)
	confHeader(req)
	if checkErr(err) {
		fmt.Println("Error creating request for join endpoint")
		return aa, err
	}
	res, err := client.Do(req)
	if checkErr(err) {
		fmt.Println("Error sending HTTP request to join endpoint")
		return aa, err
	}
	res.Body.Close()
	var msgs onesecmail.Messages
	for {
		time.Sleep(time.Second)
		msgs, err = mailClient.CheckMailbox(address)

		if checkErr(err) {
			fmt.Println("Error retrieving emails")
			return aa, err
		}
		if len(msgs) < 1 {
			continue
		}
		if msgs[0].Subject != "Sign-up to Atlas VPN" {
			continue
		}
		break
	}

	message, err := mailClient.ReadEmail(address, msgs[0].Id)

	if checkErr(err) {
		fmt.Println("Error retrieving email details by id")
		return aa, err
	}

	pattern := `Code: \*([^*]+)\*`

	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(message.TextBody)
	var code string
	if len(matches) >= 1 {
		code = matches[1]
	} else {
		fmt.Println("Error extracting verification code from email text")
		return aa, nil
	}

	tokenRes, err := codeToToken(code, client)
	if checkErr(err) {
		fmt.Println("Error getting token using code")
		return aa, err
	}

	url = "https://user.atlasvpn.com/v1/auth/confirm"
	method = "GET"
	req, err = http.NewRequest(method, url, nil)
	if checkErr(err) {
		fmt.Println("Error creating request for confirm endpoint")
		return aa, err
	}
	confHeader(req)
	req.Header.Add("Authorization", "Bearer "+tokenRes.Token)
	res, err = client.Do(req)
	if checkErr(err) {
		fmt.Println("Error sending HTTP request to confirm endpoint")
		return aa, err
	}
	d, err := bodyToBytes(res.Header.Get("Content-Encoding"), &res.Body)
	if checkErr(err) {
		fmt.Println("Error converting response body to bytes")
		return aa, err
	}
	err = json.Unmarshal(d, &tokenRes)
	if checkErr(err) {
		fmt.Println("Error decoding confirm json")
		return aa, err
	}

	if res.StatusCode != 200 {
		fmt.Printf("failed to confirm account, response status: %s", res.Status)
		return aa, nil
	}
	aa.MailAddr = address
	aa.Auth = "Bearer " + tokenRes.Token
	aa.LastEmailReadDate = message.Date

	return aa, nil
}

func extractAuthUrlFromMail(textBody string) (string, error) {
	pattern := `Complete sign-up \( (https:\/\/[^\s]+) \)`

	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(textBody)
	var code string
	if len(matches) >= 1 {
		code = matches[1]
		code = strings.ReplaceAll(code, "&amp;", "&")
		return code, nil
	} else {
		fmt.Println("Error extracting verification code from email text")
		return code, errors.New("not found match for  " + pattern)
	}

}

func waitForAuthEmails(mailClient onesecmail.Client, acc *AtlasAccount) {
	var msgs onesecmail.Messages
	var err error
	for {
		for {
			time.Sleep(time.Second)
			msgs, err = mailClient.CheckMailbox(acc.MailAddr)
			if checkErr(err) {
				fmt.Println("Error retrieving emails")
				return
			}
			if len(msgs) < 1 {
				continue
			}
			if msgs[0].Subject != "Sign-up to Atlas VPN" {
				continue
			}
			if msgs[0].Date == acc.LastEmailReadDate {
				continue
			}
			break
		}
		message, err := mailClient.ReadEmail(acc.MailAddr, msgs[0].Id)

		if checkErr(err) {
			fmt.Println("Error retrieving email details by id")
			return
		}

		acc.LastEmailReadDate = message.Date
		code, err := extractAuthUrlFromMail(message.TextBody)
		if checkErr(err) {
			fmt.Println("Error exacting auth url from mail")
			return
		}
		fmt.Printf("%s\n", code)
		go mailClient.ClearMailbox(acc.MailAddr)
	}
}

func main() {

	startTime := time.Now()
	cl := &http.Client{}
	client := onesecmail.NewClient()

	fmt.Print("[i] Creating account...\n")
	t := time.Now()
	acc, err := generateAtlasVPNAccount("", client, cl, nil)
	if checkErr(err) {
		return
	}
	fmt.Printf("[i] Created account in %s, fetching uuid...\n", time.Since(t).String())
	t = time.Now()
	uuid, err := acc.fetchUuid(cl)
	if checkErr(err) {
		return
	}
	fmt.Printf("[i] Fetched uuid in %s, starting creating accounts...\n", time.Since(t).String())
	i := 0
	tunnel := make(chan bool)
	for i < 10 {
		go generateAtlasVPNAccount(uuid, client, cl, tunnel)
		i++
	}
	fmt.Printf("[i] Waiting for accounts to create...\n")
	i = 0
	for i < 10 {
		<-tunnel
		i++
	}
	parts := strings.Split(acc.MailAddr, "@")
	fmt.Printf("\n[i] Your account email addres: %s\n", acc.MailAddr)

	fmt.Printf("[i] Registration took %s\n", time.Since(startTime).String())
	fmt.Printf("[i] To check mailbox/login next time go to https://www.1secmail.com/?login=%s&domain=%s\n", parts[0], parts[1])

	fmt.Printf("[i] Go to AtlasVpn app and enter e-mail, then the link will be printed here or just quit this app\n")
	waitForAuthEmails(*client, &acc)

}
